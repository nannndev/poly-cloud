# Poly Cloud — Landing Page

The public page for Poly Cloud. Fully static and **never calls the backend API**:
Poly Cloud's backend is self-hosted, so a public page cannot reach it.

## Development

```sh
npm install
npm run dev        # http://localhost:3100
```

Port 3100 keeps it clear of `apps/frontend`, which uses port 3000.

## Build

```sh
npm run generate   # static output in .output/public
npx serve .output/public
```

## Deploying to Vercel

Create a new Vercel project from this repo, then set:

| Setting | Value |
|---|---|
| Root Directory | `apps/landing` |
| Framework Preset | Nuxt.js |
| Build Command | `npm run generate` |
| Output Directory | `.output/public` |

`vercel.json` in this folder already carries the build command, output directory,
and security headers, so Root Directory is usually the only field you need to set.

## Pages

| Route | Contents |
|---|---|
| `/` | Hero, problem, features, providers, architecture, screenshots, install, open source |
| `/contributors` | Everyone who has contributed, read from the GitHub API at build time |
| `/support` | Donation channels and the non-financial ways to help |

## External links and donation channels

Every outbound URL lives in [`app/config/site.ts`](app/config/site.ts) — the repo
address, the GitHub links, and the donation channels. Change a username there and
the whole site follows.

Donation links are built from `REPO_OWNER`. If your payment accounts use a
different handle than your GitHub username, edit the `url` on those entries
directly.

`CRYPTO` is deliberately empty. Only fill it with an address you control and have
verified — a wrong address on a public page sends money nowhere recoverable.

## GitHub data

`/contributors` and the open-source section fetch from the GitHub REST API while
the site is being built, so the deployed output is static HTML with no client-side
requests.

Two consequences worth knowing:

- **New contributors appear on the next deploy**, not immediately.
- **The API is rate limited** to 60 requests per hour per IP for unauthenticated
  calls, and returns 404 for a private repository. Either way the build still
  succeeds: the stats block is hidden and the contributors page offers a link to
  GitHub instead. `failOnError: false` in `nuxt.config.ts` keeps a failed fetch
  from failing the deploy.

## Screenshots

Image files live in `public/shots/`. All three were scrubbed of personal details
before being committed. If you add a new one, check first that no email address,
account name, or personal browser tab was captured.

Images used by the page:

- `explorer.png` — the cross-account file list
- `accounts.png` — connected accounts and their quotas
- `connect.png` — the add-account dialog

## A note on claims

Claims on this page are kept to what actually works. The architecture section
states plainly that splitting a file across accounts is not available yet, so
nobody arrives expecting it.
