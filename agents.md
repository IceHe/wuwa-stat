# Agent Workflow

Apply this checklist at the end of every conversation that changes project files:

1. If a conversation changes project files, redeploy the affected service before finishing. For the frontend, run `npm run build` in `frontend/` and sync `frontend/dist/` to `/var/www/wuwa-stat/`. For the backend, run the relevant Go tests, build `backend/server`, and restart `wuwa-stat-backend.service`. Verify the deployment with an HTTP health check and the production site; also confirm unaffected services remain available.
2. Commit and push the conversation's changes to the current branch only after both the relevant tests and deployment succeed. If either fails, report the blocker and leave the worktree changes intact.
3. Record the change history, context, design decisions, and implementation in a relevant Markdown file under `docs/`. Update that record in the same conversation as the code change.

Keep deployment scoped to the changed service when the other service does not need a new build, but still verify that both frontend and backend remain available. Never remove or overwrite unrelated user changes while preparing a deployment or commit.
