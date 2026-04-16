# discord_bot

This service hosts the Discord bot process. It uses `discordgo` to register commands, manage the Discord session, and handle bot events.

This bot is intended to run in its own container alongside the other services.

## Interaction with other applications

The current bot code listens for completed simulation notifications via Postgres and posts results back into Discord. The API and simulation worker remain separate services, which keeps bot concerns isolated from request validation and simulation execution.
