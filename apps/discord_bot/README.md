# discord_bot

This application is responsible for the creation of the discord bot. The `discordgo` library is used to conveniently setup the bot, and add the event handlers.

This bot should be run in it's own container, isolated from the `api`.

## Interaction with other applications

This application will "route" requests our discord bot receives, to our `api` server. This is done to promote decoupling, enabling multiple front-ends to communicate with the API with ease. This decoupling will also allow us to scale these resources independently, if we ever determine that to be necessary.
