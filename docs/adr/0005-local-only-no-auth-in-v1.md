# Version 1 runs locally over stdio with no accounts, auth or hosted storage

Everything runs on the user's machine against files in their repo, so there is no account system, no auth and no secrets to manage. This keeps the first release free of the account, payment and data-residency questions. A hosted or shared mode would need its own ADR covering auth, transport and storage.
