# Webmention Support

## Client

`@TODO Write This`

## Server

Starting on a Webmention server implementation. Could eventually break this out as its own component but no energy for it right now.

### Open Questions

- Where to store webmentions for a post?
    - Jekyll-compatible data files? JSON/YAML/? https://jekyllrb.com/docs/datafiles/

### Implementation notes

- need a webmention endpoint
    - exposed in template for post detail
- verify request target (on our end) first
- receiving handler should reply 202 Accepted and return, but handle everything else async
    - verify request source (on their end)
        - reject when source == target (self-link WTH?)
        - verify that the target (local) is mentioned in the source
