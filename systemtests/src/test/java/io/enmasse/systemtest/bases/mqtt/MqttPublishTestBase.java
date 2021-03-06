/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.systemtest.bases.mqtt;

import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.eclipse.paho.client.mqttv3.IMqttClient;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.slf4j.Logger;

import io.enmasse.address.model.Address;
import io.enmasse.address.model.AddressBuilder;
import io.enmasse.systemtest.bases.TestBase;
import io.enmasse.systemtest.bases.shared.ITestBaseShared;
import io.enmasse.systemtest.logs.CustomLogger;
import io.enmasse.systemtest.model.address.AddressType;
import io.enmasse.systemtest.mqtt.MqttClientFactory;
import io.enmasse.systemtest.mqtt.MqttClientFactory.Builder;
import io.enmasse.systemtest.mqtt.MqttUtils;
import io.enmasse.systemtest.utils.AddressUtils;

public abstract class MqttPublishTestBase extends TestBase implements ITestBaseShared {

    private static final String MYTOPIC = "mytopic";
    private static final Logger log = CustomLogger.getLogger();

    public void testPublishQoS0() throws Exception {
        List<MqttMessage> messages = Stream.generate(MqttMessage::new).limit(3).collect(Collectors.toList());
        messages.forEach(m -> {
            m.setPayload(UUID.randomUUID().toString().getBytes(StandardCharsets.UTF_8));
            m.setQos(0);
        });

        this.publish(messages, 0);
    }

    public void testPublishQoS1() throws Exception {
        List<MqttMessage> messages = Stream.generate(MqttMessage::new).limit(3).collect(Collectors.toList());
        messages.forEach(m -> {
            m.setPayload(UUID.randomUUID().toString().getBytes(StandardCharsets.UTF_8));
            m.setQos(1);
        });

        this.publish(messages, 1);
    }

    public void testPublishQoS2() throws Exception {
        List<MqttMessage> messages = Stream.generate(MqttMessage::new).limit(3).collect(Collectors.toList());
        messages.forEach(m -> {
            m.setPayload(UUID.randomUUID().toString().getBytes(StandardCharsets.UTF_8));
            m.setQos(2);
        });

        this.publish(messages, 2);
    }

    public void testRetainedMessages() throws Exception {
        Address topic = new AddressBuilder()
                .withNewMetadata()
                .withNamespace(getSharedAddressSpace().getMetadata().getNamespace())
                .withName(AddressUtils.generateAddressMetadataName(getSharedAddressSpace(), "test-topic1"))
                .endMetadata()
                .withNewSpec()
                .withType("topic")
                .withAddress("test-topic1")
                .withPlan(topicPlan())
                .endSpec()
                .build();
        resourcesManager.setAddresses(topic);

        MqttMessage retainedMessage = new MqttMessage();
        retainedMessage.setQos(1);
        retainedMessage.setPayload("retained-message".getBytes(StandardCharsets.UTF_8));
        retainedMessage.setId(1);
        retainedMessage.setRetained(true);

        // send retained message to the topic
        Builder publisherBuilder = new MqttClientFactory.Builder();
        customizeClient(publisherBuilder);
        try(IMqttClient publisher = publisherBuilder.create()) {
            publisher.connect();
            publisher.publish(topic.getSpec().getAddress(), retainedMessage);
            publisher.disconnect();
        }

        // each client which will subscribe to the topic should receive retained message!
        Builder subscriberBuilder = new MqttClientFactory.Builder();
        customizeClient(subscriberBuilder);
        try(IMqttClient subscriber = subscriberBuilder.create()) {
            subscriber.connect();
            CompletableFuture<MqttMessage> messageFuture = new CompletableFuture<>();
            subscriber.subscribe(topic.getSpec().getAddress(), (topic1, message) -> messageFuture.complete(message));
            MqttMessage receivedMessage = messageFuture.get(1, TimeUnit.MINUTES);
            assertTrue(receivedMessage.isRetained(), "Retained message expected");

            subscriber.disconnect();
        }
    }

    private void publish(List<MqttMessage> messages, int subscriberQos) throws Exception {

        Address dest = new AddressBuilder()
                .withNewMetadata()
                .withNamespace(getSharedAddressSpace().getMetadata().getNamespace())
                .withName(AddressUtils.generateAddressMetadataName(getSharedAddressSpace(), MYTOPIC))
                .endMetadata()
                .withNewSpec()
                .withType("topic")
                .withAddress(MYTOPIC)
                .withPlan(topicPlan())
                .endSpec()
                .build();
        resourcesManager.setAddresses(dest);

        Builder clientBuilder = new MqttClientFactory.Builder()
                .mqttConnectionOptions(options -> {
                    options.setConnectionTimeout(options.getConnectionTimeout() * 2); // Default is 30 seconds, increase it to 1 min.
                    options.setAutomaticReconnect(true);
                });
        customizeClient(clientBuilder);

        try (IMqttClient client = clientBuilder.create()) {
            log.info("Connecting");
            client.connect();

            List<CompletableFuture<MqttMessage>> receiveFutures = MqttUtils.subscribeAndReceiveMessages(client, dest.getSpec().getAddress(), messages.size(), subscriberQos);
            List<CompletableFuture<Void>> publishFutures = MqttUtils.publish(client, dest.getSpec().getAddress(), messages);

            int publishCount = MqttUtils.awaitAndReturnCode(publishFutures, 1, TimeUnit.MINUTES);
            assertThat("Incorrect count of messages published",
                    publishCount, is(messages.size()));

            int receivedCount = MqttUtils.awaitAndReturnCode(receiveFutures, 2, TimeUnit.MINUTES);
            assertThat("Incorrect count of messages received",
                    receivedCount, is(messages.size()));
        }

    }

    protected void customizeClient(Builder mqttClientBuilder) {
        //optional to implement
    }

    protected String topicPlan() {
        return getDefaultPlan(AddressType.TOPIC);
    }

}
