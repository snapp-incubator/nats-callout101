# NATS Authentication Callout Example (Go)

This repository demonstrates the **NATS Authentication Callout** method implemented in Go.
Introduced in NATS server version v2.10, this powerful feature allows you
to implement custom authentication logic by handling authentication requests and responses
over the NATS protocol itself.

## What is NATS Callout Authentication?

Traditionally, NATS authentication often relies on configuring users and permissions directly within the NATS server's configuration file. The Callout authentication method provides a more dynamic and flexible alternative.
This mechanism actually relies on NATS core protocol, which means you need to have at least one account that uses traditional authentication mechanism and the authenticator service connects through it.
All other accounts will be redirected to authenticator using NATS request and response.

Here's how it works step by step:

1.  When a client attempts to connect to the NATS server, the server can be configured to send an authentication request message to a specific NATS subject.
2.  A dedicated service (like the one demonstrated in this repository) subscribes to this subject and receives the authentication request.
3.  This service can then perform any custom authentication logic you define (e.g., querying a database, calling an external authentication service, etc.).
4.  The service then publishes an authentication response back to the NATS server on a designated reply subject.
5.  Based on the response, the NATS server either accepts or rejects the client connection.

## Why Use Callout Authentication?

This method offers several significant advantages:

* **Dynamic User Management:** You can manage users and their permissions outside of the NATS server configuration, allowing for more dynamic and programmatic control.
* **Integration with Existing Systems:** Easily integrate NATS authentication with your existing identity management systems (e.g., LDAP, OAuth providers, internal databases).
* **Centralized Authentication Logic:** Consolidate your authentication logic into a dedicated service, making it easier to manage and audit.
* **Reduced Configuration Complexity:** For complex authentication scenarios, managing users in a separate service can be simpler than managing a large NATS configuration file.
* **Custom Authentication Schemes:** Implement authentication methods beyond basic username/password, such as token-based authentication or multi-factor authentication.

## This Repository Demonstrates:

This repository provides a basic Go implementation of a service that can act as a NATS authentication callout handler. It includes:

* **Example NATS Callout Service:** A Go program that subscribes to the authentication request subject.
* **Basic Authentication Logic:** A simple example of how you might implement your authentication logic within the service.
* **Response Publishing:** Demonstrates how to format and publish the authentication response back to the NATS server.
* **Example NATS Server Configuration:** You might also include a basic `nats.conf` file showing how to enable and configure the callout authentication feature on the NATS server.
* **Example NATS Client:** A simple NATS client that attempts to connect to the server, triggering the callout process.
