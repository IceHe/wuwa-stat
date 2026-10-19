# Shared Player ID and Account Summary

The record input tabs use one shared player ID from `App.vue`. The matching account summary is also resolved centrally and provided to every `PlayerIdField`, so all tabs show the same account details and current waveplate value.

Changing the player ID clears the previous summary while the account lookup runs. Sequence checks prevent a slower response for an earlier ID from replacing the current account. Refreshes after record submission update the shared account summary for every tab.

Validation: `npm run build` in `frontend/`; `go test ./...` in `backend/`. Production verification checks the stat frontend and backend health endpoint.
