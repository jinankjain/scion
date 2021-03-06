// Copyright 2018 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package trust defines type Store, a unified interface for TRC and Certificate
// retrieval.
package trust

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto/cert"
	"github.com/scionproto/scion/go/lib/crypto/trc"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/dedupe"
	"github.com/scionproto/scion/go/lib/infra/modules/trust/trustdb"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

const (
	// Handler lifetime
	HandlerTimeout = 3 * time.Second
)

var (
	ErrEndOfTrail           = "Reached end of trail, but no trusted TRC found"
	ErrMissingAuthoritative = "Trust store is authoritative for requested object, and object was not found"
)

// Store manages requests for TRC and Certificate Chain objects.
//
// Chain and TRC requests from the local process (running the trust store)  are
// handled by GetValidChain/GetChain and GetValidTRC/GetTRC respectively, while
// requests from other services can be handled via NewXxxReqHandler methods.
//
// By default, a Store object can only return objects that are already present
// in the database. To allow a Store to use the SCION network to retrieve
// objects from other infrastructure services, an infra.Messenger must be set
// with SetMessenger.
//
// Store is backed by a sqlite3 database in package
// go/lib/infra/modules/trust/trustdb.
type Store struct {
	mu           sync.Mutex
	trustdb      *trustdb.DB
	trcDeduper   *dedupe.Deduper
	chainDeduper *dedupe.Deduper
	config       *Config
	// local AS
	ia  addr.IA
	log log.Logger
	// ID of the last infra message that was sent out by the Store
	msgID uint64
	msger infra.Messenger
}

// NewStore initializes a TRC/Certificate Chain cache/resolver backed by db.
// Parameter local must specify the AS in which the trust store resides (which
// is used during request forwarding decisions). When sending infra messages,
// the trust store will use IDs starting from startID, and increment by one for
// each message.
func NewStore(db *trustdb.DB, local addr.IA, startID uint64, options *Config,
	logger log.Logger) (*Store, error) {

	if options == nil {
		options = &Config{}
	}
	store := &Store{
		trustdb: db,
		ia:      local,
		config:  options,
		log:     logger,
		msgID:   startID,
	}
	return store, nil
}

// SetMessenger enables network access for the trust store via msger. The
// messenger can only be set once.
func (store *Store) SetMessenger(msger infra.Messenger) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.msger != nil {
		panic("messenger already set")
	}
	store.msger = msger
	store.trcDeduper = dedupe.New(store.trcRequestFunc, 0, 0)
	store.chainDeduper = dedupe.New(store.chainRequestFunc, 0, 0)
}

// trcRequestFunc is the dedupe.RequestFunc for TRC requests.
func (store *Store) trcRequestFunc(ctx context.Context, request dedupe.Request) dedupe.Response {
	req := request.(*trcRequest)
	trcReqMsg := &cert_mgmt.TRCReq{
		ISD:       req.isd,
		Version:   req.version,
		CacheOnly: req.cacheOnly,
	}
	trcMsg, err := store.msger.GetTRC(ctx, trcReqMsg, req.server, req.id)
	if err != nil {
		return wrapErr(err)
	}
	trcObj, err := trcMsg.TRC()
	if err != nil {
		return wrapErr(common.NewBasicError("Unable to parse TRC message", err, "msg", trcMsg))
	}

	if req.version != 0 && trcObj.Version != req.version {
		return wrapErr(common.NewBasicError("Remote server responded with bad version", nil,
			"got", trcObj.Version, "expected", req.version))
	}

	if req.postHook != nil {
		return dedupe.Response{Data: trcObj, Error: req.postHook(ctx, trcObj)}
	}
	return dedupe.Response{Data: trcObj}
}

// chainRequestFunc is the dedupe.RequestFunc for Chain requests.
func (store *Store) chainRequestFunc(ctx context.Context, request dedupe.Request) dedupe.Response {
	req := request.(*chainRequest)
	chainReqMsg := &cert_mgmt.ChainReq{
		RawIA:     req.ia.IAInt(),
		Version:   req.version,
		CacheOnly: req.cacheOnly,
	}
	chainMsg, err := store.msger.GetCertChain(ctx, chainReqMsg, req.server, req.id)
	if err != nil {
		return wrapErr(common.NewBasicError("Unable to get CertChain from peer", err))
	}
	chain, err := chainMsg.Chain()
	if err != nil {
		return wrapErr(common.NewBasicError("Unable to parse CertChain message", err))
	}
	if req.version != 0 && chain.Leaf.Version != req.version {
		return wrapErr(common.NewBasicError("Remote server responded with bad version", nil,
			"got", chain.Leaf.Version, "expected", req.version))
	}
	if req.postHook != nil {
		return dedupe.Response{Data: chain, Error: req.postHook(ctx, chain)}
	}
	return dedupe.Response{Data: chain}
}

// GetValidTRC asks the trust store to return a valid TRC for isd. Trail should
// contain a sequence of cross-signing ISDs to be used during validation, with
// the requested TRC being the first one.
func (store *Store) GetValidTRC(ctx context.Context, isd addr.ISD,
	trail ...addr.ISD) (*trc.TRC, error) {

	if len(trail) > 0 && trail[0] != isd {
		return nil, common.NewBasicError(fmt.Sprintf("bad trail, should start with ISD=%d\n", isd),
			nil, "trail", trail)
	}
	// FIXME(scrye): This needs support for anycasting to remote core ISDs
	return store.getValidTRC(ctx, trail, true, &snet.Addr{IA: addr.IA{I: isd}, Host: addr.SvcCS})
}

// getValidTRC recursively follows trail to create a fully validated trust
// chain leading up to trail[0].  Given a trail composed of:
//   [ISD1, ISD2, ISD3, ISD4]
// getValidTRC first tries to see if the TRC for ISD1 is in trustdb. If it's
// not, it recursively calls getValidTRC on new trail:
//   [ISD2, ISD3, ISD4]
// and eventually:
//   [ISD3, ISD4]
// Suppose the TRC for ISD3 is in the database. The function returns the TRC
// and nil. The caller now has access to the TRC for ISD3, and needs to obtain
// the TRC for ISD2. It issues a call to the backend passing the TRC of ISD3 as
// the validator. Once it gets the TRC for ISD2, it returns it. The TRC for
// ISD2 is then used to download the TRC for ISD1.
func (store *Store) getValidTRC(ctx context.Context, trail []addr.ISD,
	recurse bool, server net.Addr) (*trc.TRC, error) {

	if len(trail) == 0 {
		// We've reached the end of the trail and did not find a trust anchor,
		// propagate this information to the caller.
		return nil, common.NewBasicError(ErrEndOfTrail, nil)
	}

	if trail[0] == 0 {
		return nil, common.NewBasicError("value 0 is not a valid ISD number", nil)
	}

	trcObj, err := store.trustdb.GetTRCVersionCtx(ctx, trail[0], 0)
	if err != nil || trcObj != nil {
		return trcObj, err
	}

	// The TRC needed to perform verification is not in trustdb; advance the
	// trail and recursively try to get the next TRC.
	nextTRC, err := store.getValidTRC(ctx, trail[1:], recurse, server)
	if err != nil {
		return nil, err
	}
	if recurse == false {
		return nil, common.NewBasicError("TRC not found in DB (valid requested), "+
			"and recursion disabled", nil, "isd", trail[0])
	}
	return store.getTRCFromNetwork(ctx, &trcRequest{
		isd:      trail[0],
		version:  0,
		id:       store.nextID(),
		server:   server,
		postHook: store.newTRCValidator(nextTRC),
	})
}

// GetTRC asks the trust store to return a TRC of the requested
// version without performing any verification. If the TRC is not available, it
// is requested from the authoritative CS.
func (store *Store) GetTRC(ctx context.Context,
	isd addr.ISD, version uint64) (*trc.TRC, error) {

	return store.getTRC(ctx, isd, version, true, nil)
}

// getTRC attempts to grab the TRC from the database; if the TRC is not found,
// it follows up with a network request (if allowed).  Parameter recurse
// specifies whether this function is allowed to create new network requests.
// Parameter requester contains the node that caused the function to be called,
// or nil if the function was called due to a local feature.
func (store *Store) getTRC(ctx context.Context, isd addr.ISD, version uint64,
	recurse bool, requester net.Addr) (*trc.TRC, error) {

	trcObj, err := store.trustdb.GetTRCVersionCtx(ctx, isd, version)
	if err != nil || trcObj != nil {
		return trcObj, err
	}
	if recurse == false {
		return nil, common.NewBasicError("TRC not found in DB, and recursion disabled", nil,
			"isd", isd, "version", version, "requester", requester)
	}
	if err := store.isLocal(requester); err != nil {
		return nil, err
	}
	server, err := store.ChooseServer(addr.IA{I: isd})
	if err != nil {
		return nil, common.NewBasicError("Error determining server to query", err,
			"requested_isd", isd, "requested_version", version)
	}
	return store.getTRCFromNetwork(ctx, &trcRequest{
		isd:      isd,
		version:  version,
		id:       store.nextID(),
		server:   server,
		postHook: nil, // Disable verification / database insertion
	})
}

func (store *Store) getTRCFromNetwork(ctx context.Context, req *trcRequest) (*trc.TRC, error) {
	responseC, cancelF := store.trcDeduper.Request(ctx, req)
	defer cancelF()
	select {
	case response := <-responseC:
		if response.Error != nil {
			return nil, response.Error
		}
		return response.Data.(*trc.TRC), nil
	case <-ctx.Done():
		return nil, common.NewBasicError("Context done while waiting for TRC",
			ctx.Err(), "isd", req.isd, "version", req.version)
	}
}

// newTRCValidator returns a TRC validation callback with validator as
// trust anchor. If validation succeeds, the validated TRC is also inserted in
// the trust database.
func (store *Store) newTRCValidator(validator *trc.TRC) ValidateTRCF {
	return func(ctx context.Context, trcObj *trc.TRC) error {
		if validator == nil {
			return common.NewBasicError("TRC verification error, nil verifier", nil,
				"target", trcObj)
		}
		if _, err := trcObj.Verify(validator); err != nil {
			return common.NewBasicError("TRC verification error", err)
		}
		if _, err := store.trustdb.InsertTRCCtx(ctx, trcObj); err != nil {
			return common.NewBasicError("Unable to store TRC in database", err)
		}
		return nil
	}
}

// GetValidChain asks the trust store to return a valid certificate chain for ia.
// Trail should contain a sequence of cross-signing ISDs to be used during
// validation, with the ISD of the certificate chain being the first one.
func (store *Store) GetValidChain(ctx context.Context, ia addr.IA,
	trail ...addr.ISD) (*cert.Chain, error) {

	if len(trail) > 0 && trail[0] != ia.I {
		return nil, common.NewBasicError(fmt.Sprintf("bad trail, should start with ISD=%d\n", ia.I),
			nil, "trail", trail)
	}

	// FIXME(scrye): Currently send message to CS in remote AS, but this should
	// change once server hints can be passed to the trust store.
	return store.getValidChain(ctx, ia, trail, true, &snet.Addr{IA: ia, Host: addr.SvcCS})
}

func (store *Store) getValidChain(ctx context.Context, ia addr.IA, trail []addr.ISD,
	recurse bool, server net.Addr) (*cert.Chain, error) {

	chain, err := store.trustdb.GetChainVersionCtx(ctx, ia, 0)
	if err != nil || chain != nil {
		return chain, err
	}
	if store.config.MustHaveLocalChain && store.ia.Eq(ia) {
		return nil, common.NewBasicError(ErrMissingAuthoritative, nil,
			"requested_ia", ia)
	}
	// Chain not found, so we'll need to fetch one. First, fetch the TRC we'll
	// need during certificate chain validation.
	trcObj, err := store.getValidTRC(ctx, trail, recurse, server)
	if err != nil {
		return nil, err
	}

	if recurse == false {
		return nil, common.NewBasicError("Chain not found in DB (valid chain requested), and "+
			"recursion disabled", nil, "ia", ia)
	}
	return store.getChainFromNetwork(ctx, &chainRequest{
		ia:       ia,
		version:  0,
		id:       store.nextID(),
		server:   server,
		postHook: store.newChainValidator(trcObj),
	})
}

// GetChain asks the trust store to return a certificate chain of
// requested version without performing any verification. If the certificate
// chain is not available, it is requested from the authoritative CS.
func (store *Store) GetChain(ctx context.Context, ia addr.IA,
	version uint64) (*cert.Chain, error) {

	return store.getChain(ctx, ia, version, true, nil)
}

// getChain attempts to grab the Certificate Chain from the database; if the
// Chain is not found, it follows up with a network request (if allowed).
// Parameter recurse specifies whether this function is allowed to create new
// network requests. Parameter client contains the node that caused the
// function to be called, or nil if the function was called due to a local
// feature.
func (store *Store) getChain(ctx context.Context, ia addr.IA, version uint64,
	recurse bool, client net.Addr) (*cert.Chain, error) {

	chain, err := store.trustdb.GetChainVersionCtx(ctx, ia, version)
	if err != nil || chain != nil {
		return chain, err
	}
	// If we're authoritative for the requested IA, error out now.
	if store.config.MustHaveLocalChain && store.ia.Eq(ia) {
		return nil, common.NewBasicError(ErrMissingAuthoritative, nil,
			"requested ia", ia)
	}
	if recurse == false {
		return nil, common.NewBasicError("Chain not found in DB, and recursion disabled", nil,
			"ia", ia, "version", version, "client", client)
	}
	if err := store.isLocal(client); err != nil {
		return nil, err
	}
	server, err := store.ChooseServer(ia)
	if err != nil {
		return nil, common.NewBasicError("Error determining server to query", err,
			"requested_ia", ia, "requested_version", version)
	}
	return store.getChainFromNetwork(ctx, &chainRequest{
		ia:       ia,
		version:  version,
		id:       store.nextID(),
		server:   server,
		postHook: nil,
	})
}

// newChainValidator returns a Chain validation callback with verifier as trust
// anchor. If validation succeeds, the certificate chain is also inserted in
// the trust database.
func (store *Store) newChainValidator(validator *trc.TRC) ValidateChainF {
	return func(ctx context.Context, chain *cert.Chain) error {
		if validator == nil {
			return common.NewBasicError("Chain verification failed, nil verifier", nil,
				"target", chain)
		}
		if err := chain.Verify(chain.Leaf.Subject, validator); err != nil {
			return common.NewBasicError("Chain verification failed", err)
		}
		_, err := store.trustdb.InsertChainCtx(ctx, chain)
		if err != nil {
			return common.NewBasicError("Unable to store CertChain in database", err)
		}
		return nil
	}
}

// issueChainRequest requests a Chain from the trust store backend.
func (store *Store) getChainFromNetwork(ctx context.Context,
	req *chainRequest) (*cert.Chain, error) {

	responseC, cancelF := store.chainDeduper.Request(ctx, req)
	defer cancelF()
	select {
	case response := <-responseC:
		if response.Error != nil {
			return nil, response.Error
		}
		return response.Data.(*cert.Chain), nil
	case <-ctx.Done():
		return nil, common.NewBasicError("Context canceled while waiting for Chain",
			nil, "ia", req.ia, "version", req.version)
	}
}

func (store *Store) nextID() uint64 {
	return atomic.AddUint64(&store.msgID, 1)
}

func (store *Store) LoadAuthoritativeTRC(dir string) error {
	fileTRC, err := trc.TRCFromDir(
		dir,
		store.ia.I,
		func(err error) {
			store.log.Warn("Error reading TRC", "err", err)
		})
	if err != nil {
		return common.NewBasicError("Unable to load TRC from directory", err)
	}

	ctx, cancelF := context.WithTimeout(context.Background(), time.Second)
	dbTRC, err := store.GetValidTRC(ctx, store.ia.I, store.ia.I)
	cancelF()
	switch {
	case err != nil && common.GetErrorMsg(err) != ErrEndOfTrail:
		// Unexpected error in trust store
		return err
	case common.GetErrorMsg(err) == ErrEndOfTrail && fileTRC == nil:
		return common.NewBasicError("No TRC found on disk or in trustdb", nil)
	case common.GetErrorMsg(err) == ErrEndOfTrail && fileTRC != nil:
		_, err := store.trustdb.InsertTRC(fileTRC)
		return err
	case err == nil && fileTRC == nil:
		// Nothing to do, no TRC to load from file but we already have one in the DB
		return nil
	default:
		// Found a TRC file on disk, and found a TRC in the DB. Check versions.
		switch {
		case fileTRC.Version > dbTRC.Version:
			_, err := store.trustdb.InsertTRC(fileTRC)
			return err
		case fileTRC.Version == dbTRC.Version:
			// Because it is the same version, check if the TRCs match
			eq, err := fileTRC.JSONEquals(dbTRC)
			if err != nil {
				return common.NewBasicError("Unable to compare TRCs", err)
			}
			if !eq {
				return common.NewBasicError("Conflicting TRCs found for same version", nil,
					"db", dbTRC, "file", fileTRC)
			}
			return nil
		default:
			// file TRC is older than DB TRC, so we just ignore it
			return nil
		}
	}
}

func (store *Store) LoadAuthoritativeChain(dir string) error {
	fileChain, err := cert.ChainFromDir(
		dir,
		store.ia,
		func(err error) {
			store.log.Warn("Error reading Chain", "err", err)
		})
	if err != nil {
		return common.NewBasicError("Unable to load Chain from directory", err)
	}

	ctx, cancelF := context.WithTimeout(context.Background(), time.Second)
	dbChain, err := store.GetValidChain(ctx, store.ia, store.ia.I)
	cancelF()
	switch {
	case err != nil && common.GetErrorMsg(err) != ErrMissingAuthoritative:
		// Unexpected error in trust store
		return err
	case common.GetErrorMsg(err) == ErrMissingAuthoritative && fileChain == nil:
		return common.NewBasicError("No chain found on disk or in trustdb", nil)
	case common.GetErrorMsg(err) == ErrMissingAuthoritative && fileChain != nil:
		_, err := store.trustdb.InsertChain(fileChain)
		return err
	case err == nil && fileChain == nil:
		// Nothing to do, no chain to load from file but we already have one in the DB
		return nil
	default:
		// Found a chain file on disk, and found a chain in the DB. Check versions.
		switch {
		case fileChain.Leaf.Version > dbChain.Leaf.Version:
			_, err := store.trustdb.InsertChain(fileChain)
			return err
		case fileChain.Leaf.Version == dbChain.Leaf.Version:
			// Because it is the same version, check if the chains match
			if !fileChain.Eq(dbChain) {
				return common.NewBasicError("Conflicting chains found for same version", nil,
					"db", dbChain, "file", fileChain)
			}
			return nil
		default:
			// file chain is older than DB chain, so we just ignore it
			return nil
		}
	}
}

// NewTRCReqHandler returns an infra.Handler for TRC requests coming from a
// peer, backed by the trust store. If recurse is set to true, the handler is
// allowed to issue new TRC requests over the network.  This method should only
// be used when servicing requests coming from remote nodes.
func (store *Store) NewTRCReqHandler(recurse bool) infra.Handler {
	f := func(r *infra.Request) {
		handler := &trcReqHandler{
			request: r,
			store:   store,
			log:     store.log,
			recurse: recurse,
		}
		handler.Handle()
	}
	return infra.HandlerFunc(f)
}

// NewChainReqHandler returns an infra.Handler for Certificate Chain
// requests coming from a peer, backed by the trust store. If recurse is set to
// true, the handler is allowed to issue new TRC and Certificate Chain requests
// over the network. This method should only be used when servicing requests
// coming from remote nodes.
func (store *Store) NewChainReqHandler(recurse bool) infra.Handler {
	f := func(r *infra.Request) {
		handler := chainReqHandler{
			request: r,
			store:   store,
			log:     store.log,
			recurse: recurse,
		}
		handler.Handle()
	}
	return infra.HandlerFunc(f)
}

// NewTRCPushHandler returns an infra.Handler for TRC pushes coming from a
// peer, backed by the trust store. TRCs are pushed by local BSes during
// beaconing. Pushes are allowed from all local AS sources.
func (store *Store) NewTRCPushHandler() infra.Handler {
	f := func(r *infra.Request) {
		handler := trcPushHandler{
			request: r,
			store:   store,
			log:     store.log,
		}
		handler.Handle()
	}
	return infra.HandlerFunc(f)
}

// NewChainPushHandler returns an infra.Handler for Certifificate Chain pushes
// coming from a peer, backed by the trust store. Certificate chains are pushed
// by other ASes during core registration. Pushes are allowed from all
// local ISD sources.
func (store *Store) NewChainPushHandler() infra.Handler {
	f := func(r *infra.Request) {
		handler := chainPushHandler{
			request: r,
			store:   store,
			log:     store.log,
		}
		handler.Handle()
	}
	return infra.HandlerFunc(f)
}

// isLocal returns an error if address is not part of the local AS (or if the
// check cannot be made).
func (store *Store) isLocal(address net.Addr) error {
	// We need to send out a network request, but only do so if we're
	// servicing a request coming from our own AS.
	if address != nil {
		switch saddr, ok := address.(*snet.Addr); {
		case !ok:
			return common.NewBasicError("Unable to determine AS of address",
				nil, "addr", address)
		case !store.ia.Eq(saddr.IA):
			return common.NewBasicError("Object not found in DB, and recursion not "+
				"allowed for clients outside AS", nil, "addr", address)
		}
	}
	return nil
}

// ChooseServer builds a CS address for crypto material regarding the
// destination AS.
func (store *Store) ChooseServer(destination addr.IA) (net.Addr, error) {
	if len(store.config.LocalCSes) != 0 {
		return store.config.LocalCSes[rand.Intn(len(store.config.LocalCSes))], nil
	}
	if destination.A == 0 {
		pathSet := snet.DefNetwork.PathResolver().Query(store.ia, addr.IA{I: destination.I})
		path := pathSet.GetAppPath("")
		if path == nil {
			return nil, common.NewBasicError("Unable to find core AS", nil)
		}
		a := &snet.Addr{IA: path.Entry.Path.DstIA(), Host: addr.SvcCS}
		return a, nil
	}
	a := &snet.Addr{IA: destination, Host: addr.SvcCS}
	return a, nil
}

// wrapErr build a dedupe.Response object containing nil data and error err.
func wrapErr(err error) dedupe.Response {
	return dedupe.Response{Error: err}
}
