brigade
=======

*an S3 bucket proxy*

Brigade gives read access to an S3 bucket as a normal HTTP service. It’s
useful for letting clients on your network access a bucket without going
through AWS IAM authentication.

Brigade respects HTTP conditional request headers (`If-None-Match` and
`If-Modified-Since` in particular), and will generate JSON directory listings
using `ListObjects` for any request whose path ends in `/`.
