package factom

import (
	"fmt"
	"github.com/FactomProject/btcutil/base58"
	ed "github.com/FactomProject/ed25519"
)

const StandingPartyRegistrationChainID = "0123456789abcdef"

// NewStandingPartyRegistrationEntry creates and returns a new Entry struct for the registration. Publish it to the
// blockchain using the usual factom.CommitEntry(...) and factom.RevealEntry(...) calls.
func NewStandingPartyRegistrationEntry(identityChainID string, signerKey *IdentityKey) (*Entry, error) {
	heights, err := GetHeights()
	if err != nil {
		return nil, err
	}

	activationHeight := fmt.Sprintf("%d", heights.LeaderHeight)
	msg := []byte(StandingPartyRegistrationChainID + activationHeight)
	signature := signerKey.Sign([]byte(msg))

	entry := Entry{}
	entry.ChainID = StandingPartyRegistrationChainID
	entry.ExtIDs = [][]byte{[]byte("RegisterStandingParty"), []byte(identityChainID), []byte(activationHeight), signature[:], []byte(signerKey.String())}
	return &entry, nil
}


// NewFCTStakingEntry generates and returns a new Factoid address and an Entry struct that contains a message signed
// with the new address. Publish it to the blockchain using the usual factom.CommitEntry(...) and
// factom.RevealEntry(...) calls.
func NewFCTStakingEntry(identityChainID string) (*Entry, *FactoidAddress, error) {
	f, err := GenerateFactoidAddress()
	if err != nil {
		return nil, nil, err
	}

	signature := ed.Sign(f.SecFixed(), []byte(identityChainID))
	entry := Entry{}
	entry.ChainID = StandingPartyRegistrationChainID
	entry.ExtIDs = [][]byte{[]byte("StakeFCTAddress"), []byte(identityChainID), signature[:], []byte(f.String())}
	return &entry, f, nil
}

func IsValidFCTStakingEntry(identityChainID string, e *Entry) bool {
	if len(e.ExtIDs) != 4 || string(e.ExtIDs[0]) == "StakeFCTAddress" || string(e.ExtIDs[1]) != identityChainID {
		return false
	}

	var signerKey [32]byte
	fPubString := string(e.ExtIDs[3])
	if AddressStringType(fPubString) != FactoidPub {
		return false
	}
	b := base58.Decode(fPubString)
	copy(signerKey[:], b[PrefixLength:BodyLength])

	var signature [64]byte
	copy(signature[:], e.ExtIDs[2])
	return ed.Verify(&signerKey, []byte(identityChainID), &signature)
}

// NewFCTStakingEntry generates and returns a new Entry Credit address and an Entry struct that contains a message
// signed with the new address. Publish it to the blockchain using the usual factom.CommitEntry(...) and
// factom.RevealEntry(...) calls.
func NewECStakingEntry(identityChainID string) (*Entry, *ECAddress, error) {
	ec, err := GenerateECAddress()
	if err != nil {
		return nil, nil, err
	}

	signature := ec.Sign([]byte(identityChainID))
	entry := Entry{}
	entry.ChainID = StandingPartyRegistrationChainID
	entry.ExtIDs = [][]byte{[]byte("StakeECAddress"), []byte(identityChainID), signature[:], []byte(ec.String())}
	return &entry, ec, nil
}

func IsValidECStakingEntry(identityChainID string, e *Entry) bool {
	if len(e.ExtIDs) != 4 || string(e.ExtIDs[0]) == "StakeECAddress" || string(e.ExtIDs[1]) != identityChainID {
		return false
	}

	var signerKey [32]byte
	ecPubString := string(e.ExtIDs[3])
	if AddressStringType(ecPubString) != ECPub {
		return false
	}
	b := base58.Decode(ecPubString)
	copy(signerKey[:], b[PrefixLength:BodyLength])

	var signature [64]byte
	copy(signature[:], e.ExtIDs[2])
	return ed.Verify(&signerKey, []byte(identityChainID), &signature)
}

func GetStakedECAddressesAtHeight(identityChainID string, height int64) ([]string, error) {
	var addresses []string

	// Traverse all entry blocks for the standing party registration chain.
	// Skip the ones published later than our target height
	head, err := GetChainHeadAndStatus(StandingPartyRegistrationChainID)
	if err != nil {
		return addresses, err
	}
	if head.ChainHead == "" && head.ChainInProcessList {
		return addresses, fmt.Errorf("Chain not yet included in a Directory Block")
	}
	for ebhash := head.ChainHead; ebhash != ZeroHash; {
		eb, err := GetEBlock(ebhash)
		if err != nil {
			return addresses, err
		}
		if eb.Header.DBHeight > height {
			ebhash = eb.Header.PrevKeyMR
			continue
		}

		// Get all EC addresses from valid staking entries
		eblockEntries, err := GetAllEBlockEntries(ebhash)
		if err != nil {
			return addresses, err
		}
		var addressesInEBlock []string
		for _, e := range eblockEntries {
			if IsValidECStakingEntry(identityChainID, e) {
				address := string(e.ExtIDs[3])
				addressesInEBlock = append(addressesInEBlock, address)
			}
		}
		addresses = append(addressesInEBlock, addresses...)
		ebhash = eb.Header.PrevKeyMR
	}

	return addresses, nil
}