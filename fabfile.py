# coding=utf-8

from fabric.api import (
    local,
)


def build():
    local(
        'gox -os="darwin windows" -arch="386 amd64"'
    )
    local(
        'gox -os="linux freebsd" -arch="386 amd64 arm"'
    )
