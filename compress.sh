#!/bin/sh
# Please install upx first, https://github.com/upx/upx/releases
for line in $(find . -iname 'annie*'); do
     upx --best "$line"
done