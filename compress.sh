#!/bin/sh
# Please install upx first, https://github.com/upx/upx/releases
find ./ -xdev -maxdepth 1 -type f -iname 'lux*' -executable -exec upx --best --brute --ultra-brute {} \;
