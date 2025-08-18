# Eventloop - Backend

> This project is a part of Tilde 4.0 HSP PES-ECC's summer mentoring program.

## Mentees
- [Lalith B Seervi](https://github.com/lalithbseervi)

## Mentors
- [Aditya Hegde](https://github.com/bwaklog)
- [Akshaj](https://github.com/unhexate)
- [Nathan](https://github.com/polarhive)
- [Tejas](https://github.com/tejas-techstack)

## Getting Started

### Check for Golang installation
First, ensure you have installed Golang using this command:
```bash
go version
```

If you haven't installed Go, you can do so by following [this guide](https://www.bytesizego.com/blog/installing-golang) or refer the [Golang's official site](https://go.dev/doc/install).

### Environment Variables

#### PRODUCTION
Set `PRODUCTION` env variable to either `true` or `false`, and ensure that you are using `true` during deployment. Incorrect usage of the env variable will break the authentication flow.

#### Database Config
Then, create an account on [Couchbase Capella](https://cloud.couchbase.com/sign-up), create a bucket named `eventloop`, and then a scope named `eventloop`, and then the collections, i.e,
-   `participants`
-   `teams`
-   `forms`
-   `form_response`
-   `events`
-   `dbAuthorisedUsers`

Under the cluster's settings, go to the `Allowed IP Addresses` tab, and use `Allow Access From Anywhere`.

Then, go to `Access Control` tab, and create a new user. Note that these are the values for `DB_USERNAME` and `DB_PASSWORD` env variables.

You can get the `DB_CONNECTION_STRING` from the `Connect` tab.

#### Google OAuth Client
Now, go to [Google Console](https://console.cloud.google.com) and click on `View all products` at the end of the page.
You should be able to see `Google Auth Platform`. Click on it and then visit the `Clients` tab.

Next, create a new client, whose `Application Type` will be `Web Application`.
Under the `Authorised Javascript origins` section, add a URI which is simply your frontend endpoint. If you're deploying on Vercel, that would be `frontendProjectName.vercel.app`.

Add a URI in the same manner under the `Authorised redirect URIs` section, except the URI would be that of your backend endpoint.
Once done, you can proceed with creating the client.

#### SECRET_KEYS
In order to set the QR secret keys, you may generate them in any way you prefer, or you can use [this tool](https://jwtsecrets.com/).


## Running the project

Next, install [Vercel CLI](https://vercel.com/docs/cli) using this command:
```bash
npm i -g vercel
```

### On local machine
```bash
vercel dev
```

### On Vercel
```bash
vercel --prod
```