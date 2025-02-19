# Puppet forge cache proxy

This is a small go gin application that servers a proxy for the forgeapi. It will download modules locally to a directory.


# building
go build

# running

    ./puppet-forge-cache-proxy -port 8081 -cacheDir ./cache

    CUSTOM_FORGE_URL="http://localhost:8080"
    # Module name to install
    MODULE_NAME="puppetlabs-stdlib"
    # Install the module from the custom Forge URL
    puppet module install $MODULE_NAME --module_repository $CUSTOM_FORGE_URL
