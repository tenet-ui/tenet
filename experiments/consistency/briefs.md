# Consistency experiment: briefs and prompt parts

Each `## brief: <id>` section is one fixed brief. The runner builds each prompt from one brief, the `addendum-tools` section (tools condition only), and the `contract` section. Do not edit these texts after the first run; the report depends on them.

## contract

Reply with one complete HTML document and nothing else. Put all CSS in one `<style>` element in the head. Use no external assets, no external fonts, no images from URLs, and no JavaScript. Do not add commentary before or after the document.

## addendum-tools

The project's design system is available through the tenet MCP tools. Before you write the page, call `get_point_of_view`, then call `get_tokens`, `get_component`, `get_template` or `get_flow` for each item that fits the brief. Build the page only from the tokens those tools return.

## brief: payments-list

Build a payments list page for a small online business. Each row shows the amount, the customer, the status and the date, and opens the payment's detail page. Include a filter control and an export button. The page has a heading and a short description under it.

## brief: empty-state

Build the empty state shown on a payments list when the business has no payments yet. It has a title, one line of supporting text, and one action that creates a payment link.

## brief: account-setup

Build the first step of account setup for a small online business. Ask for the business name and the country, and give the user a continue action and a back action.

## brief: sign-in

Build a sign-in form with an email field and a password field, a primary sign-in button, and a secondary action to create an account.

## brief: settings-card

Build a settings card with a title, a one-sentence description, and three actions: save, cancel, and delete the account.
