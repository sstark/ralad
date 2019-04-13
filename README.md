Ralad
=====

Simple http downloader. Its main feature is that the user has to agree with
each redirect interactively.

Usage
-----

    ralad [flags] url

    Flags:
    -rdisplay string
            redirect display: full|part|truncate (default "truncate")
    -rpolicy string
            set redirect confirmation policy: always|relaxed|never (default "relaxed")
    -unsafeTLS
            ignore TLS certificate errors

Ralad tries to figure out a useful output file name if the server does not
supply a good one already.

Example
-------

    ralad https://github.com/sstark/snaprd/releases/download/v1.3/snaprd-1.3-linux-x86_64.zip
    redirect to https://github-production-release-asset-2e65be.s3.amazonaws.com/1515...? (y/n) y
     1.44 MiB / 1.44 MiB ▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰ 100.00% 2.10 MiB/s 0s
    1505790 bytes written
    $ ls
    snaprd-1.3-linux-x86_64.zip

