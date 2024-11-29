# gateway

This is a gateway that front-end clients send requests to. The gateway routes all incoming requests to the appropriate backend services. In addition to routing requests the gateway also performs:

- Request authentication and authorization
- Rate-limiting
- Request validation
- Caching

## Why this exists

The primary reason this exists is to simplify supporting multiple front-end clients (web apps, Discord bot, etc), and how they interact with back-end services.

For example, consider the Discord Bot. It has unique authentication requirements, as a request from the Discord bot is issued on behalf of a Discord user. These requests need to authenticate not only the Discord bot itself, but also the Discord user that issued the request to the bot. Additionally, we may want the ability to handle these requests differently from the web app requests.

Performing all of this logic within a single REST API would be possible, but when combined with other features of Saint, the complexity of the API would increase. This approach should manage complexity by segregating the resposibility of front-end request handling and back-end service communication to a single service (the gateway).
