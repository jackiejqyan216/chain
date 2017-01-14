const Connection = require('./connection')
const AccessTokens = require('./accessTokens')
const Accounts = require('./accounts')
const Assets = require('./assets')
const Balances = require('./balances')
const Config = require('./config')
const MockHsmKeys = require('./mockHsmKeys')
const Transactions = require('./transactions')
const TransactionFeeds = require('./transactionFeeds')
const UnspentOutputs = require('./unspentOutputs')

/**
 * The Chain API Client object is the root object for all API interactions.
 * To interact with Chain Core, a Client object must always be instantiated
 * first.
 * @class
 */
class Client {
  /**
   * constructor - create a new Chain client object capable of interacting with
   * the specified Chain Core.
   *
   * @param {String} baseUrl - Chain Core URL.
   * @param {String} token - Chain Core client token for API access.
   * @returns {Client}
   */
  constructor(baseUrl, token) {
    this.connection = new Connection(baseUrl, token)

    /**
     * API actions for access tokens
     * @type {AccessTokens}
     */
    this.accessTokens = new AccessTokens(this)

    /**
     * API actions for accounts
     * @type {Accounts}
     */
    this.accounts = new Accounts(this)

    /**
     * API actions for assets.
     * @type {Assets}
     */
    this.assets = new Assets(this)

    /**
     * API actions for balances.
     * @type {Balances}
     */
    this.balances = new Balances(this)

    /**
     * API actions for config.
     * @type {Config}
     */
    this.config = new Config(this)

    /**
     * @property {MockHsmKeys} keys API actions for Mock HSM keys.
     * @property {Connection} signerConnection Mock HSM signer connection.
     */
    this.mockHsm = {
      keys: new MockHsmKeys(this),
      signerConnection: new Connection('http://localhost:1999/mockhsm')
    }

    /**
     * API actions for transactions.
     * @type {Transactions}
     */
    this.transactions = new Transactions(this)

    /**
     * API actions for transaction feeds.
     * @type {TransactionFeeds}
     */
    this.transactionFeeds = new TransactionFeeds(this)

    /**
     * API actions for unspent outputs.
     * @type {UnspentOutputs}
     */
    this.unspentOutputs = new UnspentOutputs(this)
  }


  /**
   * Submit a request to the stored Chain Core connection.
   *
   * @param {String} path
   * @param {object} [body={}]
   * @returns {Promise}
   */
  request(path, body = {}) {
    return this.connection.request(path, body)
  }
}

module.exports = Client