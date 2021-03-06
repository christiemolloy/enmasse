/*
 * Copyright 2017-2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.keycloak.spi;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.keycloak.credential.CredentialModel;
import org.mockito.ArgumentCaptor;

import java.util.HashSet;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

public class ScramPasswordHashProviderTest {
    private ScramPasswordHashProvider hashprovider;
    private CredentialModel credentialModel;

    @BeforeEach
    public void setup() {
        hashprovider = new ScramPasswordHashProvider("scramsha1", 10000, "HmacSHA1", "SHA-1");
        credentialModel = mock(CredentialModel.class);

        ArgumentCaptor<byte[]> saltCaptor = ArgumentCaptor.forClass(byte[].class);
        doAnswer(invocation -> {
            when(credentialModel.getSalt()).thenReturn(saltCaptor.getValue());
            return null;
        }).when(credentialModel).setSalt(saltCaptor.capture());

        ArgumentCaptor<String> passwordCaptor = ArgumentCaptor.forClass(String.class);
        doAnswer(invocation -> {
            when(credentialModel.getValue()).thenReturn(passwordCaptor.getValue());
            return null;
        }).when(credentialModel).setValue(passwordCaptor.capture());


        ArgumentCaptor<Integer> iterationsCaptor = ArgumentCaptor.forClass(Integer.class);
        doAnswer(invocation -> {
            when(credentialModel.getHashIterations()).thenReturn(iterationsCaptor.getValue());
            return null;
        }).when(credentialModel).setHashIterations(iterationsCaptor.capture());
    }

    @Test
    public void testProviderVerifiesCorrect() {
        hashprovider.encode("testpassword", 25000, credentialModel);
        assertTrue(hashprovider.verify("testpassword", credentialModel));
    }

    @Test
    public void testProviderFailsIncorrect() {
        hashprovider.encode("testpassword", 25000, credentialModel);
        assertFalse(hashprovider.verify("wrongpassword", credentialModel));
    }

    @Test
    public void testSameValueEncodedDifferently() {
        Set<String> encoded = new HashSet<>();
        for (int i = 0; i < 100; i++) {
            hashprovider.encode("testpassword", 25000, credentialModel);
            if (!encoded.add(credentialModel.getValue())) {
                fail("Duplicate encoding of the password");
            }
        }

    }

}
