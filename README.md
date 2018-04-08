# TrustNet Blockchain
A public Bring Your Own Identity (BYOI) system with _"DLT stack as a library"_.

## Why TrustNet?
Fundamental question:
> _"Why do we need yet another blockchain network?"_

Before we answer this question, lets re-cap what current blockchain networks offer:
* Anonymous private identities
* Promiscuous public transactions
* Applications as "second-class" citizen
* Network as the application controller

Above properties make it very difficult, if not impossible, to build **decentralized native applications** that require following:
* a strong and well known identity of the user
* application level privacy and encryption of transactions
* DLT support in native applications
* custom/exensible DLT capabilities
* applications using multiple decentralized ledgers

This is the reason TrustNet was created and designed from ground up as _"DLT stack as a library"_, making it possible to build decentralized native applications.

## What TrustNet Does?
**For Users:**
* Bring Your Own Identity (BYOI)
* Privacy (You control who sees what about your Identity)
* Security (No single point of mass vulnerability)
* Availability (Resilient to high number of node failures)
* Ownership (You decide which node on the network can be used as an Identity service!)

**For Applications:**
* A public Identity Management Network
* Private applications using strong public identities
* Native Decentralized Application with custom DLT capabilities

## How TrustNet Works?
* Abstracts p2p protocol and consensus algorithm into library
* Allows application complete control over transaction processing
* Library automatically adjusts “world state” based on consensus
```
Application (transaction business logic)
     /\
     ||
     \/
DLT Consensus Platform (protocol layer)
     /\
     ||
     \/
DLT Consensus Engine (consensus layer)
```

## DLT Comparison
|Feature|TrustNet|Ethereum|
|----|----|----|
|Objective|Identity and privacy, Native DApps|Smart contracts based DApps|
|Identity Model|Strong Public identity, BYOI IMS|Anonymous private identities|
|Privacy Model|Strong privacy, application level encryption|Public/non-private transactions|
|Application Model|Native (full control) DApp|EVM bytecode based DApp|
|DLT Stack Model|Multi-network, stack as a library|single network, stack as the controller|
