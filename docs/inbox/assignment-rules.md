# Assignment Rules

## Modes

| Mode | Behavior |
|------|----------|
| `manual` | No auto-assignment; conversations enter unassigned queue |
| `round_robin` | Rotate across active team members |
| `least_loaded` | Assign to agent with fewest open conversations |

## Flow

```
new conversation (no assignee)
        ↓
assignment engine (if mode != manual)
        ↓
assigned_to set OR unassigned queue
```

## Configuration

Stored in `team_inbox_settings.assignment_mode` (default: `manual`).

Round-robin cursor persisted in `team_inbox_settings.round_robin_cursor`.

## Fallback

When no eligible agents exist, conversation remains unassigned.
