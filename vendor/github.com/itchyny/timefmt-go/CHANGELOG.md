# Changelog
## [v0.1.3](https://github.com/itchyny/timefmt-go/compare/v0.1.2..v0.1.3) (2021-04-14)
* implement `ParseInLocation` for configuring the default location

## [v0.1.2](https://github.com/itchyny/timefmt-go/compare/v0.1.1..v0.1.2) (2021-02-22)
* implement parsing/formatting time zone offset with colons (`%:z`, `%::z`, `%:::z`)
* recognize `Z` as UTC on parsing time zone offset (`%z`)
* fix padding on formatting time zone offset (`%z`)

## [v0.1.1](https://github.com/itchyny/timefmt-go/compare/v0.1.0..v0.1.1) (2020-09-01)
* fix overflow check in 32-bit architecture

## [v0.1.0](https://github.com/itchyny/timefmt-go/compare/2c02364..v0.1.0) (2020-08-16)
* initial implementation
