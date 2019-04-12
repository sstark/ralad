Ralad
=====

Simple http downloader. Its main feature is that the user has to agree with
each redirect interactively.

Options:

  -h
      show help
  -unsafeTLS
      ignore TLS certificate errors

Ralad tries to figure out a useful output file name, if the server does not
supply a good one already.

Example
-------

    ralad https://github.com/sstark/snaprd/releases/download/v1.3/snaprd-1.3-linux-x86_64.zip
    redirect to https://github-production-release-asset-2e65be.s3.amazonaws.com/151544929/d5dc7080-da51-11e8-805b-0bab3b8d2609?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIWNJYAX4CSVEH53A%2F20190412%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20190412T223941Z&X-Amz-Expires=300&X-Amz-Signature=824a2fa6ff341ef15b14d379d717ad3c94285e9f835a7d2e4a67583bfdfeda7b&X-Amz-SignedHeaders=host&actor_id=0&response-content-disposition=attachment%3B%20filename%3Dsnaprd-1.3-linux-x86_64.zip&response-content-type=application%2Foctet-stream? (y/n) y
     1.44 MiB / 1.44 MiB ▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰ 100.00% 2.10 MiB/s 0s
    1505790 bytes written
    $ ls
    snaprd-1.3-linux-x86_64.zip

