## ICE chain Engineering Code

ICE chain is a truly fast, permissionless, secure and scalable public blockchain platform 
which is supported by hybrid consensus technology called Minerva and a global developer community. 
 
ICE chain uses hybrid consensus combining PBFT and fPoW to solve the biggest problem confronting public blockchain: 
the contradiction between decentralization and efficiency. 

ICE chain uses PBFT as fast-chain to process transactions, and leave the oversight and election of PBFT to the hands of PoW nodes. 
Besides, Icechain integrates fruitchain technology into the traditional PoW protocol to become fPoW, 
to make the chain even more decentralized and fair. 
 
ICE chain also creates a hybrid consensus incentive model and a stable gas fee mechanism to lower the cost for the developers 
and operators of DApps, and provide better infrastructure for decentralized eco-system. 

<a href="https://github.com/iceming123/go-ice/blob/master/COPYING"><img src="https://img.shields.io/badge/license-GPL%20%20Icechain-lightgrey.svg"></a>

## Building the source


Building gice requires both a Go (version 1.14 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make gice

or, to build the full suite of utilities:

    make all

The execuable command gice will be found in the `cmd` directory.

## Running gice

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://github.com/iceming123/go-ice/wiki/Command-Line-Options)), 
also you can quickly run your own gice instance with a few common parameter combos.

### Running on the IceChain main network

```
$ gice console
```

This command will:

 * Start gice with network ID `19330` in full node mode(default, can be changed with the `--syncmode` flag after version 1.1).
 * Start up Gice's built-in interactive console,
   (via the trailing `console` subcommand) through which you can invoke all official [`web3` methods](https://github.com/iceming123/go-ice/wiki/RPC-API)
   as well as Geth's own [management APIs](https://github.com/iceming123/go-ice/wiki/Management-API).
   This too is optional and if you leave it out you can always attach to an already running Gice instance
   with `gice attach`.


### Running on the ICE Chain test network

To test your contracts, you can join the test network with your node.

```
$ gice --testnet console
```

The `console` subcommand has the exact same meaning as above and they are equally useful on the
testnet too. Please see above for their explanations if you've skipped here.

Specifying the `--testnet` flag, however, will reconfigure your Geth instance a bit:

 * Test network uses different network ID `18928`
 * Instead of connecting the main ICE chain network, the client will connect to the test network, which uses testnet P2P bootnodes,  and genesis states.


### Configuration

As an alternative to passing the numerous flags to the `gice` binary, you can also pass a configuration file via:

```
$ gice --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ gice --your-favourite-flags dumpconfig
```


### Running on the ICE Chain singlenode(private) network

To start a g
instance for single node,  run it with these flags:

```
$ gice --singlenode  console
```

Specifying the `--singlenode` flag, however, will reconfigure your Geth instance a bit:

 * singlenode network uses different network ID `400`
 * Instead of connecting the main or test Icechain network, the client has no peers, and generate fast block without committee.

Which will start sending transactions periodly to this node and mining fruits and snail blocks.
