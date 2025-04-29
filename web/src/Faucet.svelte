<script>
  import { onMount } from 'svelte';
  import { getAddress } from '@ethersproject/address';
  import { CloudflareProvider } from '@ethersproject/providers';

  let input = null;
  let faucetInfo = {
    account: '0x0000000000000000000000000000000000000000',
    network: 'testnet',
    payout: 1,
    symbol: 'ETH',
    hcaptcha_sitekey: '',
    explorer_url: '',
    explorer_txPath: '',
  };

  let mounted = false;
  let hcaptchaLoaded = false;
  let feedback = null;
  let txURL = null;

  onMount(async () => {
    const res = await fetch('/api/info');
    faucetInfo = await res.json();
    mounted = true;
  });

  window.hcaptchaOnLoad = () => {
    hcaptchaLoaded = true;
  };

  $: document.title = `${capitalize(faucetInfo.network)} Faucet`;

  let widgetID;
  $: if (mounted && hcaptchaLoaded) {
    widgetID = window.hcaptcha.render('hcaptcha', {
      sitekey: faucetInfo.hcaptcha_sitekey,
    });
  }

  async function handleRequest() {
    let address = input;
    const errMsg = 'Please enter a valid address or ENS name';
    if (address === null) {
      feedback = {
        message: errMsg,
        type: 'error',
      };
      return;
    }

    if (address.endsWith('.eth')) {
      try {
        const provider = new CloudflareProvider();
        address = await provider.resolveName(address);
        if (!address) {
          feedback = {
            message: errMsg,
            type: 'error',
          };
          return;
        }
      } catch (error) {
        feedback = {
          message: errMsg,
          type: 'error',
        };
        return;
      }
    }

    try {
      address = getAddress(address);
    } catch (error) {
      feedback = {
        message: errMsg,
        type: 'error',
      };
      return;
    }

    try {
      let headers = {
        'Content-Type': 'application/json',
      };

      if (hcaptchaLoaded) {
        const { response } = await window.hcaptcha.execute(widgetID, {
          async: true,
        });
        headers['h-captcha-response'] = response;
      }

      const res = await fetch('/api/claim', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          address,
        }),
      });

      let { msg } = await res.json();
      if (msg.includes('txhash')) {
        const txHash = msg.split(' ')[1];
        txURL = `${faucetInfo.explorer_url}/${faucetInfo.explorer_txPath}/${txHash}`;
      } else {
        txURL = null;
      }

      feedback = {
        message: msg,
        type: msg.includes('exceeded') ? 'warning' : 'success',
      };
    } catch (err) {
      console.error(err);
    }
  }

  function capitalize(str) {
    const lower = str.toLowerCase();
    return str.charAt(0).toUpperCase() + lower.slice(1);
  }
</script>

<svelte:head>
  {#if mounted && faucetInfo.hcaptcha_sitekey}
    <script
      src="https://hcaptcha.com/1/api.js?onload=hcaptchaOnLoad&render=explicit"
      async
      defer
    ></script>
  {/if}
</svelte:head>

<main>
  <section class="hero is-info is-fullheight">
    <div class="hero-head">
      <nav class="navbar">
        <div class="container wrapper">
          <div class="navbar-brand">
            <a class="navbar-item" href="../..">
              <span class="icon">
                <img src="../faucet-logo.svg" alt="logo" />
              </span>
              <span><b>{faucetInfo.symbol} Faucet</b></span>
            </a>
          </div>
          <div id="navbarMenu" class="navbar-menu">
            <div class="navbar-end">
              <span class="navbar-item">
                <a
                  class="button is-white is-outlined br-0"
                  href="https://github.com/liskhq/lsk-faucet"
                >
                  <span class="icon">
                    <img src="../github-logo.svg" alt="github-logo" />
                  </span>
                  <span>View Source</span>
                </a>
              </span>
            </div>
          </div>
        </div>
      </nav>
    </div>

    <div class="hero-body">
      <div class="container has-text-centered container-position">
        <div class="column is-8 is-offset-2">
          <h1 class="title">
            Receive {faucetInfo.payout}
            {faucetInfo.symbol} per request
          </h1>
          <h2 class="subtitle">
            Serving from {faucetInfo.account}
          </h2>
          <div id="hcaptcha" data-size="invisible"></div>
          <div class="box address-box">
            <div class="field is-grouped">
              <p class="control is-expanded m-0">
                <input
                  bind:value={input}
                  class="input address-search p-12 br-0"
                  type="text"
                  placeholder="Enter your address or ENS name"
                />
              </p>
              <p class="control">
                <button
                  on:click={handleRequest}
                  class="button is-secondary text-black request-btn br-0"
                >
                  Request
                </button>
              </p>
            </div>
            {#if feedback}
              <div class="feedback">
                <span class={feedback.type}>
                  <img src={`../${feedback.type}-icon.svg`} alt="info" />
                  {#if txURL != null}
                    <a href={txURL} target="_blank">{feedback.message}</a>
                  {:else}
                    {feedback.message}
                  {/if}
                </span>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </section>
</main>

<style>
  .hero.is-info {
    background-color: #0c0c0c;
    -webkit-background-size: cover;
    -moz-background-size: cover;
    -o-background-size: cover;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
  }
  .hero .subtitle {
    margin: 0;
    padding: 24px 0 32px 0;
    line-height: 1.5;
  }
  .box {
    border-radius: 19px;
  }
  .wrapper {
    justify-content: space-between;
    max-width: none;
    padding: 24px 48px;
  }
  .navbar-item > .icon {
    margin-right: 8px;
  }
  .hero-body .title {
    margin: 0;
  }
  .address-box {
    background-color: transparent;
    padding: 0;
  }
  .address-search {
    border: 1px solid #27272a !important;
    background: #0c0c0c;
    color: #9ca3af;
  }
  .address-search::placeholder {
    color: #9ca3af;
  }
  .m-0 {
    margin: 0;
  }
  .p-12 {
    padding: 12px;
  }
  .br-0 {
    border-radius: 0px;
  }
  .is-secondary {
    background-color: #ffffff;
    opacity: 0.7;
  }
  .text-black {
    color: #000000;
  }
  .request-btn {
    border: none;
  }
  .is-grouped {
    display: flex;
    gap: 12px; 
  }
  .container-position {
    max-width: 65%;
    bottom: 80px;
  }
  .feedback {
    font-size: 16px;
    line-height: 22px;
    margin-top: 4px;
    vertical-align: middle;
  }
  .feedback span {
    display: flex;
    justify-content: center;
  }
  .feedback img {
    margin-right: 8px;
  }
  .feedback .success {
    color: #2bd67b;
    text-decoration: underline;
  }
  .feedback .warning {
    color: #fec84b;
  }
  .feedback .error {
    color: #cf462c;
  }
</style>