# coding=utf-8

from fabric.api import (
    local,
)


def build(args='-os="linux darwin windows freebsd" -arch="386 amd64"'):
    local(
        'gox {}'.format(args)
    )
