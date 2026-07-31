---
name: bk-itsm-plugin-development
description: Use when developing backend plugin or extension mechanisms in a Codex project whose root contains sibling bk-itsm, bk-itsm-apps, and bk-itsm-integrated repositories, especially infrastructure protocols, SimpleFactory registration, service implementations, plugins packages, or AppConfig discovery.
---

# BK-ITSM Plugin Development

## Scope

Use this skill only when the current Codex project root is an ITSM generated worktree or equivalent workspace that contains these sibling repositories:

- `bk-itsm`
- `bk-itsm-apps`
- `bk-itsm-integrated`

## Architecture

Use the layered plugin architecture:

- `infrastructure/`: contracts only. Define `Protocol` interfaces and factory instances. Do not put business logic here.
- `services/`: concrete implementations. Implement behavior with duck typing and pure business logic.
- `plugins/`: assembly and registration. Bind service implementations to infrastructure factories.
- `apps.py`: discovery. Import plugin modules during app startup so registration code runs.

Keep dependencies one-way:

- Infrastructure should not import services.
- Services may depend on models and standard libraries.
- Plugins may import both infrastructure factories and service implementations.

## Directory Pattern

For a component named `<component>` under `<domain>`:

```text
bk_itsm_apps/
  <app_name>/
    infrastructure/
      <domain>/
        <component>.py
    services/
      <domain>/
        <component>/
          __init__.py
          strategy_a.py
          strategy_b.py
    plugins/
      <component>.py
    apps.py
    models/
      <domain>.py
```

Follow the existing repository layout if it differs, but preserve the same responsibilities.

## Implementation Steps

### 1. Define Contract And Factory

In `infrastructure/<domain>/<component>.py`:

- Use `typing.Protocol` for the interface.
- Use `bk_itsm.packages.utils.factory.SimpleFactory` for registration.
- The protocol defines what must happen; the factory manages which implementation handles it.

```python
from typing import Protocol

from bk_itsm.packages.utils.factory import SimpleFactory


class IMyComponent(Protocol):
    """组件能力协议"""

    def execute(self, param: str) -> bool:
        ...


my_component_factory = SimpleFactory[str, IMyComponent]("my_component")
```

### 2. Implement Service Logic

In `services/<domain>/<component>/<implementation_name>.py`:

- Implement the protocol methods.
- Do not explicitly inherit from the protocol. Use duck typing.
- Keep implementation decoupled from infrastructure definitions unless type hints require local context.
- Use full type hints and Chinese docstrings/comments.

```python
class OptimizationStrategy:
    """优化策略实现"""

    def execute(self, param: str) -> bool:
        return True
```

### 3. Register Plugin

In `plugins/<component>.py`:

- Import the factory from infrastructure.
- Import concrete implementations from services.
- Register each implementation with a unique key.

```python
from my_app.infrastructure.my_domain.my_component import my_component_factory
from my_app.services.my_domain.my_component.strategy_a import OptimizationStrategy


my_component_factory.register("optimization", OptimizationStrategy)
```

### 4. Ensure Discovery

In `apps.py`:

- Ensure `AppConfig.ready()` calls an `init_plugins()` method or existing equivalent.
- Ensure `init_plugins()` imports modules under the `plugins` package so registration code executes.
- Prefer the repository's existing plugin auto-discovery helper if present.

## Business Usage

When consuming plugin behavior from domain logic, scenario, or view code:

- Prefer dependency injection and factory retrieval over direct concrete class instantiation.
- Use the factory to resolve the implementation by key.
- Use explicit errors or defaults for missing keys; do not allow ambiguous silent failures.

```python
from my_app.infrastructure.my_domain.my_component import my_component_factory


def perform_task(strategy_key: str) -> None:
    strategy_cls = my_component_factory.must_get(strategy_key)
    strategy = strategy_cls()
    strategy.execute("test")
```

## Remote Plugin Safety

When work touches existing plugin metadata and runtime plugin objects:

- Distinguish local plugins from remote API plugins by checking `api_register`.
- Check `plugin.protocol.components` for `None` before accessing local components; remote plugins may not include local component definitions.
- Use clients built by `ProtocolClientFactory` for remote plugin interaction.
- In plugin migration or sync code, explicitly filter `api_register=True` plugins when local-only handling is required.

## Completion Check

- Verify the factory has registered keys and consumers retrieve through the factory.
- Add or update tests for protocol conformance and registration where behavior changed.
- Confirm no infrastructure-to-services import cycle was introduced.
