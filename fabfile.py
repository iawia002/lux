# coding=utf-8

from fabric.api import (
    local,
)


def build():
    local(
        'gox -arch="386 amd64"'
    )
