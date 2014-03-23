// kiss-structs
package kiss

import "encoding/binary"
import "math/big"

import "io"
import "fmt"

const (
	K_HANDSHAKE_C = iota
	K_HANDSHAKE_S = iota
	K_APP_DATA    = iota
)

type KiSS_Segment struct {
	segment_type int    //one of the above enums
	raw_payload  []byte //___ENCRYPTED!!!___
}

type KiSS_HS_Client struct {
	version_number int
	public_edh_key *big.Int
}

type KiSS_HS_Server struct {
	public_dh_key  *big.Int
	public_edh_key *big.Int
}

// Returns the byte slice representation of a KiSS segment
func (segm KiSS_Segment) Bytes() []byte {
	toret := make([]byte, len(segm.raw_payload)+3)
	binary.BigEndian.PutUint16(toret[:2], uint16(len(segm.raw_payload)+1))
	toret[2] = byte(segm.segment_type)
	copy(toret[3:], segm.raw_payload)
	return toret
}

func (segm KiSS_Segment) StringRep() string {
	return fmt.Sprintf("[KiSS %d | len %d]", segm.segment_type, len(segm.raw_payload))
}

// Reads a segment from the reader
func KiSS_read_segment(rdr io.Reader) (KiSS_Segment, error) {
	//LOG(LOG_DEBUG, "of starting reading a segment")
	sgmlenbts := make([]byte, 2)
	_, err := ReadFixed(rdr, sgmlenbts)
	if err != nil {
		return KiSS_Segment{0, nil}, err
	}
	sgmlen := binary.BigEndian.Uint16(sgmlenbts)
	//LOG(LOG_DEBUG, "of readen the segment length, is of %d|%X", sgmlen, sgmlenbts)
	sgmbts := make([]byte, sgmlen)
	_, err = ReadFixed(rdr, sgmbts)
	if err != nil {
		return KiSS_Segment{0, nil}, err
	}
	//LOG(LOG_DEBUG, "KISS[%d:...]", sgmbts[0])
	return KiSS_Segment{int(sgmbts[0]), sgmbts[1:]}, nil
}

// Packs a client handshake into a byte slice
func (hsc KiSS_HS_Client) Pack() []byte {
	SASSERT(hsc.version_number == 2)
	pubkey_bytes := hsc.public_edh_key.Bytes()
	toret := make([]byte, 0)
	toret = append(toret, byte(hsc.version_number))
	toret = append(toret, pubkey_bytes...)
	return toret
}

// Packs a server handshake into a byte slice
func (hss KiSS_HS_Server) Pack() []byte {
	toret := make([]byte, 0)
	// Store DH long-term key first
	toret = append(toret, hss.public_dh_key.Bytes()...)
	// Then of storings ephemeral DH key
	toret = append(toret, hss.public_edh_key.Bytes()...)
	return toret
}

// Unpacks a client handshake into a byte slice
func KiSS_unpack_client_handshake(raw_blob []byte) (KiSS_HS_Client, error) {
	// The version number is of first byte.
	vnum := int(raw_blob[0])
	// The next 256 bytes (2048 bits) are of ephemeral DH
	pkey := big.NewInt(0).SetBytes(raw_blob[1:257])
	return KiSS_HS_Client{vnum, pkey}, nil
}

// Unpacks a server handshake into a byte slice
func KiSS_unpack_server_handshake(raw_blob []byte, checker func(*big.Int) bool) (KiSS_HS_Server, error) {
	// The first 256 bytes are the long-term DH key.
	long_dh := big.NewInt(0).SetBytes(raw_blob[:256])
	// The next 256 bytes are the short-term DH key.
	short_dh := big.NewInt(0).SetBytes(raw_blob[256:512])
	return KiSS_HS_Server{long_dh, short_dh}, nil
}
