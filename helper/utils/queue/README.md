# Messaging Topic Naming Convention (Kafka & RabbitMQ)

This document defines the **naming convention for topics and queues** across all services using **Kafka** or **RabbitMQ**. A consistent naming structure improves clarity, scalability, and maintainability across teams and environments.

---

## 🧭 Naming Format

We follow a **dot-separated structure**:

```
<domain>.<service>.<channel>.<action>
```

### Components

| Part       | Description                                                     | Example              |
|------------|------------------------------------------------------------------|----------------------|
| `domain`   | Logical grouping or business capability (DDD context)           | `communication`      |
| `service`  | Owning or responsible service name                              | `notification-svc`   |
| `channel`  | Specific resource or communication type                         | `email`, `push`      |
| `action`   | Operation being performed or published event                    | `send`, `sent`, `failed` |

---

## ✅ Naming Rules

1. Use **only lowercase letters** and **hyphens**.
2. Use **dot (`.`)** as the separator.
3. Do **not include environment** (`dev`, `prod`) in names.
4. Keep names **short, specific, and readable**.
5. Use **verbs for actions** (e.g., `send`, `created`, `failed`).
6. Define queues/topics **per communication type** if needed (e.g., email vs push).

---

## ✉️ Examples

| Purpose                                 | Name                                                  |
|-----------------------------------------|--------------------------------------------------------|
| Command: send an email                  | `communication.notification-svc.email.send`            |
| Event: email was successfully sent      | `communication.notification-svc.email.sent`            |
| Event: email sending failed             | `communication.notification-svc.email.failed`          |
| Command: send push notification         | `communication.notification-svc.push.send`             |
| Event: push notification failed         | `communication.notification-svc.push.failed`           |

---

## 🧱 Queue/Topic Types

### 🔹 Commands
Used when a service wants to trigger an action via the message queue.

Example:
```
communication.notification-svc.email.send
```

### 🔸 Events
Used when a service publishes the result of an action.

Examples:
```
communication.notification-svc.email.sent
communication.notification-svc.email.failed
```

---

## 🔄 Kafka Guidelines

- Default to **2–4 partitions** per topic.
- Example Kafka command:
  ```bash
  kafka-topics.sh --create \
    --topic communication.notification-svc.email.send \
    --partitions 4 \
    --replication-factor 1 \
    --bootstrap-server <broker-host>:9092
  ```

---

## 🐇 RabbitMQ Guidelines

- Use the same naming convention for:
    - **Exchange names**
    - **Routing keys**
    - **Queue names**
- Prefer topic exchanges for flexibility.
- Bind queues using routing patterns like:
  ```
  communication.notification-svc.email.*
  ```

Example queue name:
```
communication.notification-svc.email.send
```

---

## 📦 Multi-Channel Support

If your service handles multiple communication channels (email, push, SMS), create one topic/queue per channel:

- `communication.notification-svc.email.send`
- `communication.notification-svc.push.send`
- `communication.notification-svc.sms.send`

---

## 🔍 Discovery Tips

### Kafka
```bash
# List all topics
kafka-topics.sh --list | grep communication.
```

### RabbitMQ
- Use RabbitMQ UI or CLI to inspect exchanges/queues:
```bash
rabbitmqctl list_queues
rabbitmqctl list_exchanges
```

---

## 📌 Related

- [Kafka Topic Naming Best Practices](https://developer.confluent.io/learn/kafka-topic-naming/)
- [RabbitMQ Routing and Exchanges](https://www.rabbitmq.com/tutorials/tutorial-five-python.html)
