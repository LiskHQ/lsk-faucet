package server

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/LiskHQ/lsk-faucet/internal/chain"
	"github.com/LiskHQ/lsk-faucet/web"
)

type Server struct {
	chain.TxBuilder
	cfg *Config
}

func NewServer(builder chain.TxBuilder, cfg *Config) *Server {
	return &Server{
		TxBuilder: builder,
		cfg:       cfg,
	}
}

func (s *Server) setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(web.Dist()))
	router.Handle("/health", s.handleHealthCheck())
	limiter := NewLimiter(s.cfg.proxyCount, time.Duration(s.cfg.interval)*time.Minute)
	hcaptcha := NewCaptcha(s.cfg.hcaptchaSiteKey, s.cfg.hcaptchaSecret)
	router.Handle("/api/claim", negroni.New(limiter, hcaptcha, negroni.Wrap(s.handleClaim())))
	router.Handle("/api/info", s.handleInfo())

	return router
}

func (s *Server) Run() {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(s.setupRouter())
	log.Infof("Starting http server %d", s.cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.cfg.httpPort), n))
}

func (s *Server) handleClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		// The error always be nil since it has already been handled in limiter
		address, _ := readAddress(r)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		currBalance, err := s.GetContractInstance().BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
		if err != nil {
			log.WithError(err).Error("failed to fetch recipient balance")
			currBalance = big.NewInt(0)
		}

		txHash, err := s.TransferERC20(ctx, address, chain.LSKToWei(int64(s.cfg.payout)), currBalance)
		if err != nil {
			log.WithError(err).Error("failed to send transaction")
			renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": address,
		}).Info("Transaction sent successfully")
		resp := claimResponse{Message: fmt.Sprintf("Txhash: %s", txHash)}
		renderJSON(w, resp, http.StatusOK)
	}
}

func (s *Server) handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		renderJSON(w, infoResponse{
			Account:         s.Sender().String(),
			Network:         s.cfg.network,
			Symbol:          s.cfg.symbol,
			Payout:          strconv.Itoa(s.cfg.payout),
			HcaptchaSiteKey: s.cfg.hcaptchaSiteKey,
			ExplorerURL:     s.cfg.explorerURL,
			ExplorerTxPath:  s.cfg.explorerTxPath,
		}, http.StatusOK)
	}
}

func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		//nolint:errcheck
		w.Write([]byte("OK"))
	}
}
