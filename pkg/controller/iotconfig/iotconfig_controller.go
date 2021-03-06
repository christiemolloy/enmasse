/*
 * Copyright 2019-2020, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotconfig

import (
	"context"
	"github.com/enmasseproject/enmasse/pkg/util/iot"
	"github.com/enmasseproject/enmasse/pkg/util/loghandler"
	"github.com/pkg/errors"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"k8s.io/client-go/tools/record"
	"reflect"

	"github.com/enmasseproject/enmasse/pkg/util/cchange"

	"github.com/enmasseproject/enmasse/pkg/util/install"

	"github.com/enmasseproject/enmasse/pkg/util/recon"

	"github.com/enmasseproject/enmasse/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const ControllerName = "iotconfig-controller"

const iotPrefix = "iot.enmasse.io"
const iotServiceCaConfigMapName = "iot-service-ca"

const DeviceConnectionTypeAnnotation = iotPrefix + "/deviceConnection.type"

const RegistryTypeAnnotation = iotPrefix + "/registry.type"

var log = logf.Log.WithName("controller_iotconfig")

// Gets called by parent "init", adding as to the manager
func Add(mgr manager.Manager) error {
	name, err := iot.GetIoTConfigName()
	if err != nil {
		return errors.Wrap(err, "Failed to get IoTConfig name")
	}

	namespace, err := util.GetInfrastructureNamespace()
	if err != nil {
		return errors.Wrap(err, "Infrastructure namespace not set")
	}

	return add(mgr, newReconciler(
		mgr,
		namespace,
		name,
	))
}

func newReconciler(mgr manager.Manager, infraNamespace string, configName string) *ReconcileIoTConfig {
	return &ReconcileIoTConfig{
		client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		namespace:  infraNamespace,
		configName: configName,
		recorder:   mgr.GetEventRecorderFor(ControllerName),
	}
}

type watching struct {
	obj       runtime.Object
	openshift bool
}

func add(mgr manager.Manager, r *ReconcileIoTConfig) error {

	// Create a new controller
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IoTConfig
	err = c.Watch(&source.Kind{Type: &iotv1alpha1.IoTConfig{}}, loghandler.New(&handler.EnqueueRequestForObject{}, log, "IoTConfig"))
	if err != nil {
		return err
	}

	// watch owned objects
	ownerEventLog := log.V(2)

	for _, w := range []watching{
		{&appsv1.Deployment{}, false},
		{&appsv1.StatefulSet{}, false},
		{&corev1.Service{}, false},
		{&corev1.ConfigMap{}, false},
		{&corev1.Secret{}, false},
		{&corev1.PersistentVolumeClaim{}, false},

		{&routev1.Route{}, true},
	} {

		if w.openshift && !util.IsOpenshift() {
			// requires openshift, but we don't have it
			continue
		}

		err = c.Watch(&source.Kind{Type: w.obj}, loghandler.New(&handler.EnqueueRequestForOwner{
			OwnerType:    &iotv1alpha1.IoTConfig{},
			IsController: true,
		}, ownerEventLog, w.obj.GetObjectKind().GroupVersionKind().Kind))
		if err != nil {
			return err
		}

	}

	// watch secrets referenced by infrastructure

	s := NewSecretHandler(r.client, r.configName)
	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, s); err != nil {
		return err
	}

	// done

	return nil
}

// ensure we are implementing the interface
var _ reconcile.Reconciler = &ReconcileIoTConfig{}

type ReconcileIoTConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder

	// The name of the configuration we are watching
	// we are watching only one config, in our own namespace
	configName string
	// The name of the infrastructure namespace
	namespace string
}

// Reconcile
//
// returning an error will get the request re-queued
func (r *ReconcileIoTConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IoTConfig")

	// Get config
	config := &iotv1alpha1.IoTConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, config)

	ctx := context.TODO()

	if err != nil {

		if apierrors.IsNotFound(err) {

			reqLogger.Info("Config was not found")

			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if config.Name != r.configName {
		return r.failWrongConfigName(ctx, config)
	}

	rc := &recon.ReconcileContext{}

	original := config.DeepCopy()

	// update and store credentials

	rc.Process(func() (result reconcile.Result, e error) {
		return r.processGeneratedCredentials(ctx, config)
	})

	if rc.Error() != nil || rc.NeedRequeue() {
		log.Info("Re-queue after processing finalizers")
		return rc.Result()
	}

	// start normal reconcile

	qdrProxyConfigCtx := cchange.NewRecorder()

	rc.Process(func() (reconcile.Result, error) {
		return r.processQdrProxyConfig(ctx, config, qdrProxyConfigCtx)
	})
	rc.ProcessSimple(func() error {
		return r.processInterServiceCAConfigMap(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processServiceMesh(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processAuthService(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processTenantService(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processDeviceConnection(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processDeviceRegistry(ctx, config)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processHttpAdapter(ctx, config, qdrProxyConfigCtx)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processMqttAdapter(ctx, config, qdrProxyConfigCtx)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processSigfoxAdapter(ctx, config, qdrProxyConfigCtx)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processLoraWanAdapter(ctx, config, qdrProxyConfigCtx)
	})
	rc.Process(func() (reconcile.Result, error) {
		return r.processMonitoring(ctx, config)
	})

	return r.updateFinalStatus(ctx, original, config, rc)
}

func (r *ReconcileIoTConfig) updateStatus(ctx context.Context, original *iotv1alpha1.IoTConfig, config *iotv1alpha1.IoTConfig, err error) error {

	// update the status condition

	r.updateConditions(ctx, config, err)

	// record state change in event log

	if config.Status.Phase != original.Status.Phase {
		switch config.Status.Phase {
		case iotv1alpha1.ConfigPhaseActive:
			// record event that the infrastructure was successfully activated
			r.recorder.Eventf(config, corev1.EventTypeNormal, "Activated", "IoT infrastructure successfully activated")
		case iotv1alpha1.ConfigPhaseFailed:
			// record event that the infrastructure failed
			r.recorder.Eventf(config, corev1.EventTypeWarning, "Failed", "IoT infrastructure failed")
		}
	}

	// update status

	if !reflect.DeepEqual(config, original) {
		log.Info("IoTConfig changed, updating status")
		return r.client.Status().Update(ctx, config)
	}

	// no update needed

	return nil
}

func (r *ReconcileIoTConfig) updateFinalStatus(ctx context.Context, original *iotv1alpha1.IoTConfig, config *iotv1alpha1.IoTConfig, rc *recon.ReconcileContext) (reconcile.Result, error) {

	// do a status update

	err := r.updateStatus(ctx, original, config, rc.Error())

	// check if error is non-recoverable

	if rc.Error() != nil && util.OnlyNonRecoverableErrors(rc.Error()) {
		r.recorder.Eventf(config, corev1.EventTypeWarning, "NonRecoverableError", "Configuration failed: %v", rc.Error())
		log.Error(rc.Error(), "All non-recoverable errors. Stop trying to reconcile.")
		// strip away error, only return update error
		return rc.PlainResult(), err
	} else {
		// return result ... including status update error
		if err != nil {
			rc.AddResult(reconcile.Result{}, err)
		}
		return rc.Result()
	}
}

func (r *ReconcileIoTConfig) failWrongConfigName(ctx context.Context, config *iotv1alpha1.IoTConfig) (reconcile.Result, error) {

	config.Status.Phase = iotv1alpha1.ConfigPhaseFailed
	config.Status.Message = "The name of the resource must be 'default'"

	config.Status.GetConfigCondition(iotv1alpha1.ConfigConditionTypeReady).SetStatusError("WrongName", config.Status.Message)

	if err := r.client.Status().Update(ctx, config); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIoTConfig) processDeployment(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, deployment *appsv1.Deployment) error) error {

	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{Namespace: config.Namespace, Name: name},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, deployment, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		if err := controllerutil.SetControllerReference(config, deployment, r.scheme); err != nil {
			return err
		}

		return manipulator(config, deployment)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processStatefulSet(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, deployment *appsv1.StatefulSet) error) error {

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{Namespace: config.Namespace, Name: name},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, statefulSet, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, statefulSet, func() error {
		if err := controllerutil.SetControllerReference(config, statefulSet, r.scheme); err != nil {
			return err
		}

		return manipulator(config, statefulSet)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processService(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, service *corev1.Service) error) error {

	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{Namespace: config.Namespace, Name: name},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, service, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		if err := controllerutil.SetControllerReference(config, service, r.scheme); err != nil {
			return err
		}

		return manipulator(config, service)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processConfigMap(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, configMap *corev1.ConfigMap) error) error {

	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      name,
		},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, cm, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, cm, func() error {
		if err := controllerutil.SetControllerReference(config, cm, r.scheme); err != nil {
			return err
		}

		return manipulator(config, cm)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processSecret(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, secret *corev1.Secret) error) error {

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      name,
		},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, secret, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, secret, func() error {
		if err := controllerutil.SetControllerReference(config, secret, r.scheme); err != nil {
			return err
		}

		return manipulator(config, secret)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processPersistentVolumeClaim(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, service *corev1.PersistentVolumeClaim) error) error {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      name,
		},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, pvc, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, pvc, func() error {
		if err := controllerutil.SetControllerReference(config, pvc, r.scheme); err != nil {
			return err
		}

		return manipulator(config, pvc)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processPrometheusRule(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, manipulator func(config *iotv1alpha1.IoTConfig, rule *promv1.PrometheusRule) error) error {

	pr := &promv1.PrometheusRule{
		ObjectMeta: v1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      name,
		},
	}

	if delete {
		return install.DeleteIgnoreNotFound(ctx, r.client, pr, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, pr, func() error {
		if err := controllerutil.SetControllerReference(config, pr, r.scheme); err != nil {
			return err
		}

		return manipulator(config, pr)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processRoute(ctx context.Context, name string, config *iotv1alpha1.IoTConfig, delete bool, endpointStatus *iotv1alpha1.EndpointStatus, manipulator func(config *iotv1alpha1.IoTConfig, service *routev1.Route, endpointStatus *iotv1alpha1.EndpointStatus) error) error {

	route := &routev1.Route{
		ObjectMeta: v1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      name,
		},
	}

	if delete {
		endpointStatus.URI = ""
		return install.DeleteIgnoreNotFound(ctx, r.client, route, client.PropagationPolicy(v1.DeletePropagationForeground))
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, route, func() error {
		if err := controllerutil.SetControllerReference(config, route, r.scheme); err != nil {
			return err
		}

		return manipulator(config, route, endpointStatus)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return err
	}

	return nil
}

func (r *ReconcileIoTConfig) processGeneratedCredentials(ctx context.Context, config *iotv1alpha1.IoTConfig) (reconcile.Result, error) {

	if config.Status.Services == nil {
		config.Status.Services = make(map[string]iotv1alpha1.ServiceStatus)
	}

	original := config.DeepCopy()

	// generate auth service PSK

	if config.Status.AuthenticationServicePSK == nil {
		s, err := util.GeneratePassword(128)
		if err != nil {
			return reconcile.Result{}, err
		}
		config.Status.AuthenticationServicePSK = &s
	}

	// ensure we have all adapter status entries we want

	config.Status.Adapters = ensureAdapterStatus(config.Status.Adapters)

	// generate adapter users

	if err := initConfigStatus(config); err != nil {
		return reconcile.Result{}, err
	}

	// compare

	if !reflect.DeepEqual(original, config) {
		log.Info("Credentials change detected. Updating status and re-queuing.")
		return reconcile.Result{Requeue: true}, r.updateStatus(ctx, original, config, nil)
	} else {
		return reconcile.Result{}, nil
	}

}
