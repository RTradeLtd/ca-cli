package nacl

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/RTradeLtd/ca-cli/errs"
	"github.com/RTradeLtd/ca-cli/utils"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/nacl/secretbox"
)

func secretboxCommand() cli.Command {
	return cli.Command{
		Name:      "secretbox",
		Usage:     "encrypt and authenticate small messages using secret-key cryptography",
		UsageText: "step crypto nacl secretbox <subcommand> [arguments] [global-flags] [subcommand-flags]",
		Description: `**step crypto nacl secretbox** command group uses secret-key cryptography to
encrypt, decrypt and authenticate messages. The implementation is based on NaCl's
crypto_secretbox function.

NaCl crypto_secretbox is designed to meet the standard notions of privacy and
authenticity for a secret-key authenticated-encryption scheme using nonces. For
formal definitions see, e.g., Bellare and Namprempre, "Authenticated encryption:
relations among notions and analysis of the generic composition paradigm,"
Lecture Notes in Computer Science 1976 (2000), 531–545,
http://www-cse.ucsd.edu/~mihir/papers/oem.html. Note that the length is not
hidden. Note also that it is the caller's responsibility to ensure the
uniqueness of nonces—for example, by using nonce 1 for the first message, nonce
2 for the second message, etc. Nonces are long enough that randomly generated
nonces have negligible risk of collision.

NaCl crypto_secretbox is crypto_secretbox_xsalsa20poly1305, a particular
combination of Salsa20 and Poly1305 specified in "Cryptography in NaCl". This
function is conjectured to meet the standard notions of privacy and
authenticity.

These commands are interoperable with NaCl: https://nacl.cr.yp.to/secretbox.html

## EXAMPLES

Encrypt a message using a 256-bit secret key, a new nacl box private key can
be used as the secret:
'''
$ step crypto nacl secretbox seal nonce secretbox.key
Please enter text to seal: ********
o2NJTsIJsk0dl4epiBwS1mM4xFED7iE

$ cat message.txt | step crypto nacl secretbox seal nonce secretbox.key
o2NJTsIJsk0dl4epiBwS1mM4xFED7iE
'''

Decrypt and authenticate the message:
'''
$ echo o2NJTsIJsk0dl4epiBwS1mM4xFED7iE | step crypto nacl secretbox open nonce secretbox.key
message
'''`,
		Subcommands: cli.Commands{
			secretboxOpenCommand(),
			secretboxSealCommand(),
		},
	}
}

func secretboxOpenCommand() cli.Command {
	return cli.Command{
		Name:   "open",
		Action: cli.ActionFunc(secretboxOpenAction),
		Usage:  "authenticate and decrypt a box produced by seal",
		UsageText: `**step crypto nacl secretbox open** <nonce> <key-file>
		[--raw]`,
		Description: `**step crypto nacl secretbox open** verifies and decrypts a ciphertext using a
secret key and a nonce.

This command uses an implementation of NaCl's crypto_secretbox_open function.

For examples, see **step help crypto nacl secretbox**.`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "raw",
				Usage: "Indicates that input is not base64 encoded",
			},
		},
	}
}

func secretboxSealCommand() cli.Command {
	return cli.Command{
		Name:   "seal",
		Action: cli.ActionFunc(secretboxSealAction),
		Usage:  "produce an encrypted ciphertext",
		UsageText: `**step crypto nacl secretbox seal** <nonce> <key-file>
		[--raw]`,
		Description: `**step crypto nacl secretbox seal** encrypts and authenticates a message using
a secret key and a nonce.

This command uses an implementation of NaCl's crypto_secretbox function.

For examples, see **step help crypto nacl secretbox**.`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "raw",
				Usage: "Do not base64 encode output",
			},
		},
	}
}

func secretboxOpenAction(ctx *cli.Context) error {
	if err := errs.NumberOfArguments(ctx, 2); err != nil {
		return err
	}

	args := ctx.Args()
	nonce, keyFile := []byte(args[0]), args[1]

	if len(nonce) > 24 {
		return errors.New("nonce cannot be longer than 24 bytes")
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return errs.FileError(err, keyFile)
	} else if len(key) != 32 {
		return errors.New("invalid key file: key size is not 32 bytes")
	}

	input, err := utils.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "error reading input")
	}

	var rawInput []byte
	if ctx.Bool("raw") {
		rawInput = input
	} else {
		// DecodeLen returns the maximum length,
		// Decode will return the actual length.
		rawInput = make([]byte, b64Encoder.DecodedLen(len(input)))
		n, err := b64Encoder.Decode(rawInput, input)
		if err != nil {
			return errors.Wrap(err, "error decoding base64 input")
		}
		rawInput = rawInput[:n]
	}

	var n [24]byte
	var k [32]byte
	copy(n[:], nonce)
	copy(k[:], key)

	// Fixme: if we prepend the nonce in the seal we can use use rawInput[24:]
	// as the message and rawInput[:24] as the nonce instead of requiring one.
	raw, ok := secretbox.Open(nil, rawInput, &n, &k)
	if !ok {
		return errors.New("error authenticating or decrypting input")
	}

	os.Stdout.Write(raw)
	return nil
}

func secretboxSealAction(ctx *cli.Context) error {
	if err := errs.NumberOfArguments(ctx, 2); err != nil {
		return err
	}

	args := ctx.Args()
	nonce, keyFile := []byte(args[0]), args[1]

	if len(nonce) > 24 {
		return errors.New("nonce cannot be longer than 24 bytes")
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return errs.FileError(err, keyFile)
	} else if len(key) != 32 {
		return errors.New("invalid key: key size is not 32 bytes")
	}

	input, err := utils.ReadInput("Please enter text to seal")
	if err != nil {
		return errors.Wrap(err, "error reading input")
	}

	var n [24]byte
	var k [32]byte
	copy(n[:], nonce)
	copy(k[:], key)

	// Fixme: we can prepend nonce[:] so it's not necessary in the open.
	raw := secretbox.Seal(nil, input, &n, &k)
	if ctx.Bool("raw") {
		os.Stdout.Write(raw)
	} else {
		fmt.Println(b64Encoder.EncodeToString(raw))
	}

	return nil
}
