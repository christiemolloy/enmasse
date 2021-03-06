/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.systemtest.manager;

import static io.enmasse.systemtest.bases.iot.ITestIoTBase.IOT_PROJECT_NAMESPACE;

import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

import org.junit.jupiter.api.extension.ExtensionContext;
import org.slf4j.Logger;

import io.enmasse.address.model.AddressSpace;
import io.enmasse.iot.model.v1.IoTConfig;
import io.enmasse.iot.model.v1.IoTProject;
import io.enmasse.systemtest.UserCredentials;
import io.enmasse.systemtest.amqp.AmqpClientFactory;
import io.enmasse.systemtest.logs.CustomLogger;
import io.enmasse.systemtest.platform.apps.SystemtestsKubernetesApps;
import io.enmasse.systemtest.utils.IoTUtils;
import io.enmasse.systemtest.utils.TestUtils;
import io.enmasse.systemtest.utils.ThrowingCallable;

public class IsolatedIoTManager extends ResourceManager {

    private Logger LOGGER = CustomLogger.getLogger();
    protected AmqpClientFactory amqpClientFactory;
    protected List<IoTProject> ioTProjects;
    protected List<IoTConfig> ioTConfigs;
    private static IsolatedIoTManager instance = null;
    private UserCredentials defaultCredentials = environment.getDefaultCredentials();

    private IsolatedIoTManager() {
        ioTProjects = new ArrayList<>();
        ioTConfigs = new ArrayList<>();
    }

    public static synchronized IsolatedIoTManager getInstance() {
        if (instance == null) {
            instance = new IsolatedIoTManager();
        }
        return instance;
    }

    public void initFactories(AddressSpace addressSpace) {
        amqpClientFactory = new AmqpClientFactory(addressSpace, defaultCredentials);
    }

    public void initFactories(IoTProject project) {
        String addSpaceName = project.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName();
        this.initFactories(kubernetes.getAddressSpaceClient(project.getMetadata()
                .getNamespace()).withName(addSpaceName).get());
    }

    @Override
    public void setup() {
        kubernetes.createNamespace(IOT_PROJECT_NAMESPACE);
    }

    private static Throwable cleanup(ThrowingCallable callable, Throwable e) {
        try {
            callable.call();
            return e;
        } catch (Throwable e1) {
            if (e == null) {
                return e1;
            } else {
                e.addSuppressed(e1);
                return e;
            }
        }
    }

    @Override
    public void tearDown(ExtensionContext context) throws Exception {

        if (environment.skipCleanup()) {

            LOGGER.info("Skip cleanup is set, no cleanup process");

        } else {

            // the next lines, using cleanup(...) prevent that the failure of
            // one cleanup step prevents the next step from being executed.
            // Also we must handle Throwables here, since some parts of the code
            // call 'assert*', which throws a 'AssertionFailedError', which is
            // is an 'Error' rather than an 'Exception'.

            Throwable e = null;
            e = cleanup(() -> tearDownProjects(), e);
            e = cleanup(() -> tearDownConfigs(), e);
            if (context.getExecutionException().isPresent()) {
                final Path path = TestUtils.getFailedTestLogsPath(context);
                e = cleanup(() -> SystemtestsKubernetesApps.collectInfinispanServerLogs(path), e);
            }
            e = cleanup(() -> SystemtestsKubernetesApps.deleteInfinispanServer(), e);
            e = cleanup(() -> SystemtestsKubernetesApps.deletePostgresqlServer(), e);
            e = cleanup(() -> SystemtestsKubernetesApps.deleteH2Server(), e);

            if (e != null) {
                LOGGER.error("Error tearing down IoT test: {}", e);
                if (e instanceof Exception) {
                    throw (Exception) e;
                } else {
                    throw new Exception(e);
                }
            }

        }
    }

    private void tearDownProjects() throws Exception {
        LOGGER.info("All IoTProjects will be removed");
        for (IoTProject project : ioTProjects) {
            var iotProjectApiClient = kubernetes.getIoTProjectClient(project.getMetadata().getNamespace());
            if (iotProjectApiClient.withName(project.getMetadata().getName()).get() != null) {
                IoTUtils.deleteIoTProjectAndWait(kubernetes, project);
            } else {
                LOGGER.info("IoTProject '{}' doesn't exists!", project.getMetadata().getName());
            }
        }
        ioTProjects.clear();
    }

    private void tearDownConfigs() throws Exception {
        // delete configs
        LOGGER.info("All IoTConfigs will be removed");
        var iotConfigApiClient = kubernetes.getIoTConfigClient();
        for (IoTConfig config : ioTConfigs) {
            if (iotConfigApiClient.withName(config.getMetadata().getName()).get() != null) {
                IoTUtils.deleteIoTConfigAndWait(kubernetes, config);
            } else {
                LOGGER.info("IoTConfig '{}' doesn't exists!", config.getMetadata().getName());
            }
        }
        ioTConfigs.clear();
    }

    @Override
    public AmqpClientFactory getAmqpClientFactory() {
        return amqpClientFactory;
    }

    @Override
    public void setAmqpClientFactory(AmqpClientFactory amqpClientFactory) {
        this.amqpClientFactory = amqpClientFactory;
    }

    public void createIoTProject(IoTProject project) throws Exception {
        ioTProjects.add(project);
        IoTUtils.createIoTProject(project);
        initFactories(project);
    }

    public void deleteIoTProject(IoTProject project) throws Exception {
        IoTUtils.deleteIoTProjectAndWait(kubernetes, project);
        ioTProjects.remove(project);
    }

    public void createIoTConfig(IoTConfig ioTConfig) throws Exception {
        ioTConfigs.add(ioTConfig);
        IoTUtils.createIoTConfig(ioTConfig);
    }

    public void deleteIoTConfig(IoTConfig ioTConfig) throws Exception {
        IoTUtils.deleteIoTConfigAndWait(kubernetes, ioTConfig);
        ioTConfigs.add(ioTConfig);
    }

    public List<IoTProject> getIoTProjects() {
        return ioTProjects;
    }

    public String getTenantId() {
        return IoTUtils.getTenantId(ioTProjects.get(0));
    }

    public List<IoTConfig> getIoTConfigs() {
        return ioTConfigs;
    }

}
