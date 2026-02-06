# HelixGo

HelixGo is a Go backend framework built for long-term evolution.

It starts as a minimal REST API core and gradually evolves into a
multi-interface backend platform that supports REST APIs, Telegram Bots,
Webhooks, Web3 systems, and Exchange-style architectures.

HelixGo follows a **helix evolution philosophy**:
each new layer adds capabilities without breaking or rewriting existing systems.

---

## Why HelixGo

Most backend frameworks are either:
- too heavy at the beginning, or
- too limited for long-term growth.

HelixGo is designed to grow **with real business complexity**.

You can start with a simple REST API today, and evolve the same codebase
to support bots, event-driven systems, blockchain integrations, and
high-throughput trading services in the future.

---

## Design Philosophy

- **Minimal first, powerful later**
- **Clear architecture over framework magic**
- **Business evolution over short-term productivity**
- **REST, Bot, Webhook are first-class citizens**
- **Friendly to PHP, Java, and Go engineers**

---

## What HelixGo Is

- A backend foundation for multi-business systems
- A long-term platform, not a quick admin template
- A framework that embraces gradual architectural evolution
- A practical choice for real-world systems, not demos

---

## What HelixGo Is Not

- Not an Admin-only framework
- Not a code generator
- Not a microservice framework by default
- Not opinionated about cloud vendors or infrastructure

---

## Architecture Overview

HelixGo is built around a modular and evolvable architecture.

```text
helixgo
├─ helix-api        // REST API core (starting point)
├─ helix-bot        // Telegram / Bot interfaces
├─ helix-event      // Event, message, and webhook handling
├─ helix-admin      // Admin UI backend (optional)
├─ helix-web3       // Blockchain / wallet / signing systems
├─ helix-exchange   // Trading, matching, risk control (future)
└─ helix-core       // Shared core libraries

---

## Author & Maintainer

HelixGo is created and maintained by **magei**.

Website: https://www.hu88.cn