--------------
Key Management
--------------

Besides providing us a client that can interact with the blockchain,
we also have a keybase at our disposal to manage our private keys.
When you provide an argument to ``yarn cli``, it will load the keys
from the given database. You can also do this programmatically.
It will read/write from any
`levelup <https://www.npmjs.com/package/levelup>`__
compatible storage, so we can use
`leveldown <https://www.npmjs.com/package/leveldown>`__ in the cli,
`memdown <https://www.npmjs.com/package/memdown>`__ in test cases,
and `level.js <https://github.com/level/level.js>`__ in the browser,
and `asyncstorage <https://github.com/tradle/asyncstorage-down>`__
for react-native.

``yarn cli demo-keys.db``

The underlying code for this command:

.. code:: javascript

    const weave = require("weave");
    const leveldown = require("leveldown");
    function loadKeybase(file) {
        return weave.openDB(leveldown(file))
            .then(db => weave.KeyBase.setup(db));
    }
    let keys = await loadKeyBase("demo-keys.db");

Creating Key Pairs
------------------

Once we have access to a KeyBase, we can create key pairs.
All operations are syncronous in memory, you must explicitly call
the asynchronous function (using .then() / await) to save the keys,
otherwise all new keys are lost when you close the store.

.. code:: javascript

    // create a new key
    let demo = keys.add('demo')
    // try to create a conflicting name throws an error
    keys.add('demo')
    // instead, we can load an existing key
    let demo2 = keys.get('demo')
    // check the address
    demo.address()
    demo.address() === demo2.address()
    // add a second key and see the contents of the keybase
    let rcpt = keys.add('rcpt')
    keys.list()
    // now save them both so we can use them later
    await keys.save()

Signing and Verifying
---------------------

We will later use the keys to sign and verify transactions.
This is done as part of higher-level functions, but to get an
idea of how the signatures work, try the following. Note that
every signature include a chainID to tie it to one blockchain
(in the case of a fork), and a sequence number (for replay
protection). Both of these must be match in the verify
function for it to be considered valid.

We need the private/secret key to sign the message, but only
need the public key to verify the signature. Currently we only
handle key pairs, but when we enable importing public keys of known
contacts, we could easily use this to verify signatures of any
known contact, whose public key we have.

.. code:: javascript

    // create two different messages
    let msg = Buffer.from("my secret message")
    let msg2 = Buffer.from("modified message")
    // chainID is used to tie this transaction to one blockchain
    let chainID = 'proper-chain'
    // sign the msg with the demo key, get signature and sequence
    let {sig, seq} = demo.sign(msg, chainID)
    // verifies with all proper
    demo.verify(msg, sig, chainID, seq)
    // changing key, msg, chain, or sequence invalidates sig
    rcpt.verify(msg, sig, chainID, seq)
    demo.verify(msg2, sig, chainID, seq)
    demo.verify(msg, sig, 'fork-chain', seq)
    demo.verify(msg, sig, chainID, 10)
