/*
 * Copyright 2019-2020, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.systemtest.iot.isolated;

import static io.enmasse.systemtest.TestTag.SMOKE;
import static io.enmasse.systemtest.time.TimeoutBudget.ofDuration;
import static java.time.Duration.ofMinutes;
import static java.util.Collections.singletonMap;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.enmasse.iot.model.v1.CommonAdapterContainersBuilder;
import io.enmasse.iot.model.v1.ContainerConfigBuilder;
import io.enmasse.iot.model.v1.IoTConfig;
import io.enmasse.iot.model.v1.JavaContainerConfigBuilder;
import io.enmasse.systemtest.Environment;
import io.enmasse.systemtest.bases.TestBase;
import io.enmasse.systemtest.bases.iot.ITestIoTIsolated;
import io.enmasse.systemtest.condition.Kubernetes;
import io.enmasse.systemtest.iot.DefaultDeviceRegistry;
import io.enmasse.systemtest.iot.IoTTestSession;
import io.enmasse.systemtest.platform.KubeCMDClient;
import io.enmasse.systemtest.platform.apps.SystemtestsKubernetesApps;
import io.enmasse.systemtest.utils.IoTUtils;
import io.enmasse.systemtest.utils.TestUtils;
import io.fabric8.kubernetes.api.model.Quantity;

@Tag(SMOKE)
@Kubernetes
class SimpleK8sDeployTest extends TestBase implements ITestIoTIsolated {

    private static final Logger log = LoggerFactory.getLogger(SimpleK8sDeployTest.class);
    private static final String NAMESPACE = Environment.getInstance().namespace();
    private static IoTConfig config;
    private io.enmasse.systemtest.platform.Kubernetes client = io.enmasse.systemtest.platform.Kubernetes.getInstance();

    @BeforeAll
    static void setup() throws Exception {
        Map<String, String> secrets = new HashMap<>();
        secrets.put("iot-auth-service", "systemtests-iot-auth-service-tls");
        secrets.put("iot-tenant-service", "systemtests-iot-tenant-service-tls");
        secrets.put("iot-device-connection", "systemtests-iot-device-connection-tls");
        secrets.put("iot-device-registry", "systemtests-iot-device-registry-tls");

        var r1 = new ContainerConfigBuilder()
                .withNewResources().addToLimits("memory", new Quantity("64Mi")).endResources()
                .build();
        var j2 = new JavaContainerConfigBuilder()
                .withNewContainerConfig()
                .withNewResources().addToLimits("memory", new Quantity("256Mi")).endResources()
                .endContainerConfig()
                .build();

        var commonContainers = new CommonAdapterContainersBuilder()
                .withNewAdapterLike(j2).endAdapter()
                .withNewProxyLike(r1).endProxy()
                .withNewProxyConfiguratorLike(r1).endProxyConfigurator()
                .build();

        var jdbcEndpoint = SystemtestsKubernetesApps.deployPostgresqlServer();

        config = IoTTestSession.createDefaultConfig()

                .editOrNewSpec()

                .editOrNewAdapters()

                .editOrNewHttp()
                .withNewContainersLike(commonContainers).endContainers()
                .endHttp()

                .editOrNewMqtt()
                .withNewContainersLike(commonContainers).endContainers()
                .endMqtt()

                .editOrNewSigfox()
                .withNewContainersLike(commonContainers).endContainers()
                .endSigfox()

                .editOrNewLoraWan()
                .withNewContainersLike(commonContainers).endContainers()
                .endLoraWan()

                .endAdapters()

                .withNewServices()

                .withNewAuthentication()
                .withNewContainerLike(j2).endContainer()
                .endAuthentication()

                .withNewTenant()
                .withNewContainerLike(j2).endContainer()
                .endTenant()

                .withDeviceConnection(DefaultDeviceRegistry.newPostgresBasedConnection(jdbcEndpoint))
                .withDeviceRegistry(DefaultDeviceRegistry.newPostgresBasedRegistry(jdbcEndpoint, false))

                .editDeviceConnection()
                .editJdbc()
                .editOrNewCommonServiceConfig()
                .withNewContainerLike(j2).endContainer()
                .endCommonServiceConfig()
                .endJdbc()
                .endDeviceConnection()

                .editDeviceRegistry()
                .editJdbc()
                .editServer()
                .editExternal()
                .editManagement()
                .editOrNewCommonConfig()
                .withNewContainerLike(j2).endContainer()
                .endCommonConfig()
                .endManagement()
                .endExternal()
                .endServer()
                .endJdbc()
                .endDeviceRegistry()

                .endServices()

                .endSpec()
                .build();

        final Path configTempFile = Files.createTempFile("iot-config", "json");
        try {
            Files.write(configTempFile, new ObjectMapper().writeValueAsBytes(config));
            KubeCMDClient.createFromFile(NAMESPACE, configTempFile);
        } finally {
            Files.deleteIfExists(configTempFile);
        }
    }

    @AfterAll
    static void cleanup() throws Exception {
        KubeCMDClient.deleteIoTConfig(NAMESPACE, "default");
        log.info("Waiting for IoT components to be removed");
        TestUtils.waitForNReplicas(0, NAMESPACE, singletonMap("component", "iot"), ofDuration(ofMinutes(5)));
    }

    @Test
    void testDeploy() throws Exception {
        IoTUtils.waitForIoTConfigReady(client, config);
    }

}
