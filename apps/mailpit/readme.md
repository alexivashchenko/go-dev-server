Documentation:
  https://github.com/axllent/mailpit
  https://mailpit.axllent.org/docs/

Usage:
  mailpit [flags]
  mailpit [command]

Available Commands:
  dump        Dump all messages from a database to a directory
  ingest      Ingest a file or folder of emails for testing
  readyz      Run a healthcheck to test if Mailpit is running
  reindex     Reindex the database
  sendmail    A sendmail command replacement for Mailpit
  version     Display the current version & update information

Flags:
  -d, --database string                  Database to store persistent data
      --label string                     Optional label identify this Mailpit instance
      --tenant-id string                 Database tenant ID to isolate data
  -m, --max int                          Max number of messages to store (default 500)
      --max-age string                   Max age of messages in either (h)ours or (d)ays (eg: 3d)
      --use-message-dates                Use message dates as the received dates
      --ignore-duplicate-ids             Ignore duplicate messages (by Message-Id)
      --log-file string                  Log output to file instead of stdout
  -q, --quiet                            Quiet logging (errors only)
  -v, --verbose                          Verbose logging
  -l, --listen string                    HTTP bind interface & port for UI (default "[::]:8025")
      --webroot string                   Set the webroot for web UI & API (default "/")
      --ui-auth-file string              A password file for web UI & API authentication
      --ui-tls-cert string               TLS certificate for web UI (HTTPS) - requires ui-tls-key
      --ui-tls-key string                TLS key for web UI (HTTPS) - requires ui-tls-cert
      --api-cors string                  Set API CORS Access-Control-Allow-Origin header
      --block-remote-css-and-fonts       Block access to remote CSS & fonts
      --enable-spamassassin string       Enable integration with SpamAssassin
      --allow-untrusted-tls              Do not verify HTTPS certificates (link checker & screenshots)
  -s, --smtp string                      SMTP bind interface and port (default "[::]:1025")
      --smtp-auth-file string            A password file for SMTP authentication
      --smtp-auth-accept-any             Accept any SMTP username and password, including none
      --smtp-tls-cert string             TLS certificate for SMTP (STARTTLS) - requires smtp-tls-key
      --smtp-tls-key string              TLS key for SMTP (STARTTLS) - requires smtp-tls-cert
      --smtp-require-starttls            Require SMTP client use STARTTLS
      --smtp-require-tls                 Require client use SSL/TLS
      --smtp-auth-allow-insecure         Allow insecure PLAIN & LOGIN SMTP authentication
      --smtp-strict-rfc-headers          Return SMTP error if message headers contain <CR><CR><LF>
      --smtp-max-recipients int          Maximum SMTP recipients allowed (default 100)
      --smtp-allowed-recipients string   Only allow SMTP recipients matching a regular expression (default allow all)
      --smtp-disable-rdns                Disable SMTP reverse DNS lookups
      --smtp-relay-config string         SMTP relay configuration file to allow releasing messages
      --smtp-relay-all                   Auto-relay all new messages via external SMTP server (caution!)
      --smtp-relay-matching string       Auto-relay new messages to only matching recipients (regular expression)
      --smtp-forward-config string       SMTP forwarding configuration file for all messages
      --enable-chaos                     Enable Chaos functionality (API / web UI)
      --chaos-triggers string            Enable Chaos & set the triggers for SMTP server
      --pop3 string                      POP3 server bind interface and port (default "[::]:1110")
      --pop3-auth-file string            A password file for POP3 server authentication (enables POP3 server)
      --pop3-tls-cert string             Optional TLS certificate for POP3 server - requires pop3-tls-key
      --pop3-tls-key string              Optional TLS key for POP3 server - requires pop3-tls-cert
  -t, --tag string                       Tag new messages matching filters
      --tags-config string               Load tags filters from yaml configuration file
      --tags-title-case                  TitleCase new tags generated from plus-addresses and X-Tags
      --tags-disable string              Disable auto-tagging, comma separated (eg: plus-addresses,x-tags)
      --webhook-url string               Send a webhook request for new messages
      --webhook-limit int                Limit webhook requests per second (default 1)