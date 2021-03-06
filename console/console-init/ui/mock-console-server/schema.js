/*
 * Copyright 2020, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 *
 */

const { gql } = require('apollo-server');

const typeDefs = gql`
    scalar Date

    type KeyValue {
        key: String!
        value: String!
    }

    enum AddressSpaceType {
        standard
        brokered
    }

    enum AddressType {
        queue
        topic
        subscription
        multicast
        anycast
    }

    enum AuthenticationServiceType {
        none
        standard
    }

    enum LinkRole {
        sender,
        receiver
    }

    enum MetricType {
        gauge
        counter
    }

    enum Protocol {
        amqp,
        amqps
    }

    enum EndpointServiceType_enmasse_io_v1beta1 {
        messaging
        mqtt @deprecated(reason: "deprecated"),
        console @deprecated(reason: "deprecated"),
    }

    enum CertificateProviderType_enmasse_io_v1beta1 {
        wildcard,
        certBundle,
        openshift,
        selfsigned
    }

    enum ExposeType_enmasse_io_v1beta1 {
        route
        loadbalancer
    }

    enum RouteServicePort_enmasse_io_v1beta1 {
        amqps
        https
        secure_mqtt @deprecated(reason: "deprecated")
    }

    enum RouteTlsTermination_enmasse_io_v1beta1 {
        passthrough
        reencrypt
    }

    enum MessagingEndpointType_enmasse_io_v1beta2  {
        cluster,
        nodePort,
        loadBalancer,
        route,
        ingress
    }

    enum MessagingEndpointProtocol_enmasse_io_v1beta2  {
        amqp,
        amqps,
        amqp_ws,
        amqp_wss,
    }

    enum MessagingEndpointPhase_enmasse_io_v1beta2  {
        configuring,
        active,
        terminating,
    }

    type Metric_consoleapi_enmasse_io_v1beta1 {
        name: String!
        type: MetricType!
        value: Float!
        units: String!
    }

    type AddressSpaceType_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AddressTypeSpec_consoleapi_enmasse_io_v1beta1!
    }

    type AddressSpaceTypeSpec_consoleapi_enmasse_io_v1beta1 {
        addressSpaceType: AddressSpaceType!
        displayName: String!
        longDescription: String!
        shortDescription: String!
        displayOrder: Int!
    }

    type AddressType_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AddressTypeSpec_consoleapi_enmasse_io_v1beta1!
    }

    type AddressTypeSpec_consoleapi_enmasse_io_v1beta1 {
        addressSpaceType: AddressSpaceType!
        displayName: String!
        longDescription: String!
        shortDescription: String!
        displayOrder: Int!
    }

    type AuthenticationService_admin_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AuthenticationServiceSpec_admin_enmasse_io_v1beta1!
        status: AuthenticationServiceStatus_admin_enmasse_io_v1beta1!
    }
    type AuthenticationServiceStatus_admin_enmasse_io_v1beta1 {
        host: String!
        port: Int!
    }
    type AuthenticationServiceSpec_admin_enmasse_io_v1beta1 {
        type: AuthenticationServiceType!
    }
    type AddressSpaceSchema_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AddressSpaceSchemaSpec_enmasse_io_v1beta1!
    }

    type AddressSpaceSchemaSpec_enmasse_io_v1beta1 {
        authenticationServices: [String!]
        description: String
        routeServicePorts: [RouteServicePortDescription_enmasse_io_v1beta1!]!
        certificateProviderTypes: [CertificateProviderTypeDescription_enmasse_io_v1beta1!]!
        endpointExposeTypes: [EndpointExposeTypeDescription_enmasse_io_v1beta1!]!
    }

    type RouteServicePortDescription_enmasse_io_v1beta1 {
        name: RouteServicePort_enmasse_io_v1beta1!
        displayName: String!
        routeTlsTerminations: [RouteTlsTermination_enmasse_io_v1beta1!]!
    }

    type CertificateProviderTypeDescription_enmasse_io_v1beta1 {
        name: CertificateProviderType_enmasse_io_v1beta1!
        displayName: String!
        description: String!
    }

    type EndpointExposeTypeDescription_enmasse_io_v1beta1 {
        name: ExposeType_enmasse_io_v1beta1!
        displayName: String!
        description: String!
    }

    type Connection_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: ConnectionSpec_consoleapi_enmasse_io_v1beta1!

        metrics: [Metric_consoleapi_enmasse_io_v1beta1!]!,
        links(first: Int, offset: Int, filter: String, orderBy: String): LinkQueryResult_consoleapi_enmasse_io_v1beta1!
    }

    type ConnectionSpec_consoleapi_enmasse_io_v1beta1 {
        addressSpace: AddressSpace_consoleapi_enmasse_io_v1beta1!
        hostname: String!
        containerId: String!
        protocol: Protocol!
        encrypted: Boolean!
        properties: [KeyValue!]!
        principal: String!
    }

    type Link_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: LinkSpec_consoleapi_enmasse_io_v1beta1!
        metrics: [Metric_consoleapi_enmasse_io_v1beta1!]!,
    }

    type LinkSpec_consoleapi_enmasse_io_v1beta1 {
        connection: Connection_consoleapi_enmasse_io_v1beta1!
        address: String!
        role: LinkRole!
        metrics: [Metric_consoleapi_enmasse_io_v1beta1!]!
    }

    #
    #  Types used to facilitate the paginated model queries
    #

    type AddressSpaceQueryResult_consoleapi_enmasse_io_v1beta1 {
        total: Int!
        addressSpaces: [AddressSpace_consoleapi_enmasse_io_v1beta1!]!
    }

    type AddressQueryResult_consoleapi_enmasse_io_v1beta1 {
        total: Int!
        addresses: [Address_consoleapi_enmasse_io_v1beta1!]!
    }

    type ConnectionQueryResult_consoleapi_enmasse_io_v1beta1 {
        total: Int!
        connections: [Connection_consoleapi_enmasse_io_v1beta1!]!
    }

    type LinkQueryResult_consoleapi_enmasse_io_v1beta1 {
        total: Int!
        links: [Link_consoleapi_enmasse_io_v1beta1!]!
    }

    type MessagingEndpointQueryResult_consoleapi_enmasse_io_v1beta1 {
        total: Int!
        messagingEndpoints: [MessagingEndpoint_enmasse_io_v1beta2!]!
    }

    #
    # Mirrors of Kubernetes types.  These follow the names and structure of the underlying
    # Kubernetes object exactly.  We don't need to expose every field, just the ones that
    # are important to the GraphQL interface.
    #
    # It is also possible to map types into GraphQL types (enums, other types etc) as is
    # done below for the address.spec.plan and type fields.
    #

    type AddressSpace_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AddressSpaceSpec_enmasse_io_v1beta1!
        status: AddressSpaceStatus_enmasse_io_v1beta1
        connections(first: Int, offset: Int, filter: String, orderBy: String): ConnectionQueryResult_consoleapi_enmasse_io_v1beta1!
        addresses(first: Int, offset: Int, filter: String, orderBy: String): AddressQueryResult_consoleapi_enmasse_io_v1beta1!
        metrics: [Metric_consoleapi_enmasse_io_v1beta1!]
    }

    type AddressSpaceSpec_enmasse_io_v1beta1 {
        plan:      AddressSpacePlan_admin_enmasse_io_v1beta2!
        type:      AddressSpaceType!
  
        authenticationService:  AuthenticationService_enmasse_io_v1beta1
        endpoints:              [EndpointSpec_enmasse_io_v1beta1!]
    }

    type EndpointSpec_enmasse_io_v1beta1 {
        name:        String!
        service:     EndpointServiceType_enmasse_io_v1beta1!
        certificate: CertificateSpec_enmasse_io_v1beta1
        expose:      ExposeSpec_enmasse_io_v1beta1
    }

    type CertificateSpec_enmasse_io_v1beta1 {
        provider:   CertificateProviderType_enmasse_io_v1beta1!
        secretName: String
        tlsCert:    String
        tlsKey:     String
    }

    type ExposeSpec_enmasse_io_v1beta1 {
        type:  ExposeType_enmasse_io_v1beta1!

        routeHost: String
        routeServicePort: RouteServicePort_enmasse_io_v1beta1,
        routeTlsTermination: RouteTlsTermination_enmasse_io_v1beta1,

        loadBalancerPorts: [String!]
        loadBalancerSourceRanges: [String!]
    }

    type MessagingEndpoint_enmasse_io_v1beta2 {
        metadata: ObjectMeta_v1!
        spec: MessagingEndpointSpec_enmasse_io_v1beta2!
        status: MessagingEndpointStatus_enmasse_io_v1beta2
    }

    type MessagingEndpointSpec_enmasse_io_v1beta2 {
        protocols: [MessagingEndpointProtocol_enmasse_io_v1beta2!]!
    }

    type MessagingEndpointStatus_enmasse_io_v1beta2 {
        phase: MessagingEndpointPhase_enmasse_io_v1beta2!
        type: MessagingEndpointType_enmasse_io_v1beta2!
        message: String
        host: String

        ports: [MessagingEndpointPort_enmasse_io_v1beta2!]!
        internalPorts: [MessagingEndpointPort_enmasse_io_v1beta2!]!
    }

    type MessagingEndpointPort_enmasse_io_v1beta2  {
        name: String!
        protocol: MessagingEndpointProtocol_enmasse_io_v1beta2!
        port: Int!
    }

    type AuthenticationService_enmasse_io_v1beta1 {
        name: String!
    }

    type AddressSpaceStatus_enmasse_io_v1beta1 {
        isReady: Boolean!
        messages: [String!]
        phase: String!
        caCertificate: String,
        endpointStatus: [EndpointStatus_enmasse_io_v1beta1!]!
    }

    type EndpointStatus_enmasse_io_v1beta1 {
        name: String!
        certificate: String
        serviceHost: String!
        servicePorts: [Port_enmasse_io_v1beta1!]!

        externalHost: String
        externalPorts: [Port_enmasse_io_v1beta1!]
    }

    type Port_enmasse_io_v1beta1 {
        name: String!
        port: Int!
    }

    type AddressSpec_enmasse_io_v1beta1 {
        address:      String!
        addressSpace: String!
        type:         AddressType!
        plan:         AddressPlan_admin_enmasse_io_v1beta2!
        topic:        String
    }

    type AddressStatus_enmasse_io_v1beta1 {
        isReady: Boolean!
        messages: [String!]
        phase: String!
        planStatus: AddressPlanStatus_enmasse_io_v1beta1
    }

    type AddressPlanStatus_enmasse_io_v1beta1 {
        name: String!
        partitions: Int!
    }

    type Address_consoleapi_enmasse_io_v1beta1 {
        metadata: ObjectMeta_v1!
        spec: AddressSpec_enmasse_io_v1beta1!
        status: AddressStatus_enmasse_io_v1beta1

        links(first: Int, offset: Int, filter: String, orderBy: String ): LinkQueryResult_consoleapi_enmasse_io_v1beta1!
        metrics: [Metric_consoleapi_enmasse_io_v1beta1!]
    }

    type AddressPlan_admin_enmasse_io_v1beta2 {
        metadata: ObjectMeta_v1!
        spec: AddressPlanSpec_admin_enmasse_io_v1beta2!
    }

    type AddressPlanSpec_admin_enmasse_io_v1beta2 {
        addressType: AddressType!
        displayName: String!
        longDescription: String!
        shortDescription: String!
        displayOrder: Int!
    }

    type AddressSpacePlan_admin_enmasse_io_v1beta2 {
        metadata: ObjectMeta_v1!
        spec: AddressSpacePlanSpec_admin_enmasse_io_v1beta2!
    }

    type AddressSpacePlanSpec_admin_enmasse_io_v1beta2 {
        addressPlans: [AddressPlan_admin_enmasse_io_v1beta2!]!
        addressSpaceType: AddressSpaceType,
        displayName: String!
        longDescription: String!
        shortDescription: String!
        displayOrder: Int!
    }


    type ObjectMeta_v1 {
        annotations: [KeyValue!]!
        name: String!
        namespace: String!
        resourceVersion: String!
        creationTimestamp: Date!
        uid: ID!
    }

    type User_v1 {
        metadata: ObjectMeta_v1!
        identities: [String!]!
        groups: [String!]!
        fullName: String!
    }

    type Namespace_v1 {
        metadata: ObjectMeta_v1!
        status: NamespaceStatus_v1!
    }

    type NamespaceStatus_v1 {
        phase: String!
    }

    type Query {
        hello: String

        "Returns the address spaces type defined by the system (DEPRECATED)"
        addressSpaceTypes: [AddressSpaceType!]!
        "Returns the address spaces type defined by the system optionally filtereing for a single address space type"
        addressSpaceTypes_v2: [AddressSpaceType_consoleapi_enmasse_io_v1beta1!]!

        "Returns the address types defined by the system (DEPRECATED)"
        addressTypes: [AddressType!]!
        "Returns the address types defined by the system optionally filtereing for a single address space type"
        addressTypes_v2(addressSpaceType: AddressSpaceType): [AddressType_consoleapi_enmasse_io_v1beta1!]!

        "Returns the address spaces plans defined by the system optionally filtereing for a single address space type"
        addressSpacePlans(addressSpaceType: AddressSpaceType): [AddressSpacePlan_admin_enmasse_io_v1beta2!]!

        "Returns the address plans defined by the system optionally filtering those for a matching address space plan and/or address type"
        addressPlans(addressSpacePlan: String, addressType: AddressType): [AddressPlan_admin_enmasse_io_v1beta2!]!

        "Returns the authenticationServices"
        authenticationServices: [AuthenticationService_admin_enmasse_io_v1beta1!]!
        "Returns the addressSpaceSchema"
        addressSpaceSchema: [AddressSpaceSchema_enmasse_io_v1beta1!]!
        "Returns the addressSpaceSchema optionally filtering those for a matching address space type"
        addressSpaceSchema_v2(
          addressSpaceType: AddressSpaceType
        ): [AddressSpaceSchema_enmasse_io_v1beta1!]!

        "Returns the current logged on user"
        whoami: User_v1!
        "Returns the namespaces visible to this user"
        namespaces : [Namespace_v1!]!

        "Returns the address spaces visible to this user,  optionally filtering"
        addressSpaces(first: Int, offset: Int, filter: String, orderBy: String): AddressSpaceQueryResult_consoleapi_enmasse_io_v1beta1

        "Returns the addresses visible to this user,  optionally filtering"
        addresses(first: Int, offset: Int, filter: String, orderBy: String): AddressQueryResult_consoleapi_enmasse_io_v1beta1

        "Returns the connections visible to this user,  optionally filtering"
        connections(first: Int, offset: Int, filter: String, orderBy: String): ConnectionQueryResult_consoleapi_enmasse_io_v1beta1

        "Returns the messaging certificate chain for the address space identifed by input, PEM format, suitable to be offered as a download to the user"
        messagingCertificateChain(input: ObjectMeta_v1_Input!): String!

        "Returns the command-line that, if executed, would create the given address space"
        addressSpaceCommand(input: AddressSpace_enmasse_io_v1beta1_Input!): String!

        "Returns the command-line command, if executed, would create the given address"
        addressCommand(input: Address_enmasse_io_v1beta1_Input!, addressSpace: String): String!

        "Returns the messaging endpoints for the given address space"
        messagingEndpoints(first: Int, offset: Int, filter: String, orderBy: String): MessagingEndpointQueryResult_consoleapi_enmasse_io_v1beta1

    }

    #
    # Inputs Types
    #

    input ObjectMeta_v1_Input {
        name: String
        namespace: String!
        resourceVersion: String
    }

    input AddressSpaceSpec_enmasse_io_v1beta1_Input {
        type:         String!
        plan:         String!
        
        authenticationService:  AuthenticationService_enmasse_io_v1beta1_Input
        endpoints:    [EndpointSpec_enmasse_io_v1beta1_Input!]
    }
    
    input AuthenticationService_enmasse_io_v1beta1_Input {
        name:  String!
    }

    input AddressSpace_enmasse_io_v1beta1_Input {
        metadata: ObjectMeta_v1_Input
        spec: AddressSpaceSpec_enmasse_io_v1beta1_Input
    }

    input AddressSpec_enmasse_io_v1beta1_Input {
        address:      String!
        addressSpace: String
        type:         String!
        plan:         String!
        topic:        String
    }

    input EndpointSpec_enmasse_io_v1beta1_Input {
        name:        String!
        service:     EndpointServiceType_enmasse_io_v1beta1!
        certificate: CertificateSpec_enmasse_io_v1beta1_Input
        expose:      ExposeSpec_enmasse_io_v1beta1_Input
    }

    input CertificateSpec_enmasse_io_v1beta1_Input {
        provider:   CertificateProviderType_enmasse_io_v1beta1!
        secretName: String
        tlsCert:    String
        tlsKey:     String
    }

    input ExposeSpec_enmasse_io_v1beta1_Input {
        type:  String!

        routeHost: String
        routeServicePort: RouteServicePort_enmasse_io_v1beta1,
        routeTlsTermination: RouteTlsTermination_enmasse_io_v1beta1,

        loadBalancerPorts: [String!]
        loadBalancerSourceRanges: [String!]
    }

    input Address_enmasse_io_v1beta1_Input {
        metadata: ObjectMeta_v1_Input
        spec: AddressSpec_enmasse_io_v1beta1_Input
    }

    type Mutation {
        createAddressSpace(input: AddressSpace_enmasse_io_v1beta1_Input!): ObjectMeta_v1!
        patchAddressSpace(input: ObjectMeta_v1_Input!, jsonPatch: String!, patchType : String!): Boolean
        deleteAddressSpace(input: ObjectMeta_v1_Input!): Boolean @deprecated
        "deletes addressspace(s)"
        deleteAddressSpaces(input: [ObjectMeta_v1_Input!]!): Boolean

        createAddress(input: Address_enmasse_io_v1beta1_Input!, addressSpace: String): ObjectMeta_v1!
        patchAddress(input: ObjectMeta_v1_Input!, jsonPatch: String!, patchType : String!): Boolean
        deleteAddress(input: ObjectMeta_v1_Input!): Boolean @deprecated
        "deletes addresss(es)"
        deleteAddresses(input: [ObjectMeta_v1_Input!]!): Boolean
        purgeAddress(input: ObjectMeta_v1_Input!): Boolean @deprecated
        "purges address(es)"
        purgeAddresses(input: [ObjectMeta_v1_Input!]!): Boolean

        closeConnections(input: [ObjectMeta_v1_Input!]!): Boolean
    }
`;

module.exports = typeDefs;
