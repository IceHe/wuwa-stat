# Token Retention During Network Failures

## Background

The login token is a long, highly random string stored in browser `localStorage`. Before this change, both startup session restoration and manual login cleared the token for every failed request. A temporary network outage or timeout therefore erased the saved token and forced the user to enter it again.

## Design

- Preserve the saved token when the request has no HTTP response, times out, or returns a service error.
- Clear the token only when the API explicitly returns `401`, which identifies an invalid or expired credential.
- Keep the token visible in the login input after a connection failure.
- Retry session restoration automatically when the browser emits its `online` event.
- Keep explicit logout behavior unchanged: logout clears the saved token.

## Implementation

`frontend/src/App.vue` distinguishes an explicit `401` from other request failures. Startup restoration and manual login both keep the stored token on non-401 failures and show a network warning. The online event retries restoration using the retained token. The existing API response interceptor continues to clear credentials on `401` responses from any API endpoint.

## Change History

- 2026-10-08: Fixed token loss during network failures; added automatic retry after connectivity returns.
