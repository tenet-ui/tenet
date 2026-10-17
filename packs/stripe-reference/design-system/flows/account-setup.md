# Account setup

The steps a new user takes to activate an account before going live. Each step has one primary action, and each step lists its states and transitions.

1. **Create an account.** States: `signed-out`, `account-created`. Transition: submitting the sign-up form moves `signed-out` to `account-created`. Next: step 2.
2. **Work in a sandbox.** States: `sandbox`. Transition: none; the sandbox is available as soon as the account exists. Next: step 3.
3. **Complete the account checklist.** States: `checklist-open`, `checklist-done`. Transition: each completed item moves the checklist forward; the last item moves it to `checklist-done`. Next: step 4.
4. **Verify the business.** States: `unverified`, `pending-review`, `verified`. Transition: submitting business details moves `unverified` to `pending-review`; approval moves it to `verified`. Live mode is locked until `verified`. Next: step 5.
5. **Set public details.** States: `details-incomplete`, `details-complete`. Transition: saving name, URL, support contact and statement descriptor moves it to `details-complete`. Next: step 6.
6. **Secure the account.** States: `unsecured`, `secured`. Transition: enabling two-factor authentication moves it to `secured`; inviting team members is optional at this step. End state: `secured`.

Source: docs/research/stripe-design-system.md section 7 (account setup and Dashboard onboarding).
