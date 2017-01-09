const shared = require('./shared')

/**
 * @class
 */
class HsmSigner {
  constructor() {
    this.signers = {}
  }

  addKey(key, connection) {
    const id = `${connection.baseUrl}-${connection.token || 'noauth'}`
    let signer = this.signers[id]
    if (!signer) {
      signer = this.signers[id] = {
        connection: connection,
        xpubs: []
      }
    }

    signer.xpubs.push(typeof key == 'string' ? key : key.xpub)
  }

  sign(template, cb) {
    let promise = Promise.resolve(template)

    if (Object.keys(this.signers).length == 0) {
      return shared.tryCallback(promise, cb)
    }

    for (let signerId in this.signers) {
      const signer = this.signers[signerId]

      promise = promise.then(nextTemplate =>
        signer.connection.request('/sign-transaction', {
          transactions: [nextTemplate],
          xpubs: signer.xpubs
        })
      ).then(resp => resp[0])
    }

    return shared.tryCallback(promise, cb)
  }

  signBatch(templates, cb) {
    templates = templates.filter((template) => template != null)

    let promise = Promise.resolve(templates)

    if (Object.keys(this.signers).length == 0) {
      return shared.tryCallback(promise.then(() => templates), cb)
    }

    for (let signerId in this.signers) {
      const signer = this.signers[signerId]

      promise = promise.then(nextTemplates =>
        signer.connection.request('/sign-transaction', {
          transactions: nextTemplates,
          xpubs: signer.xpubs
        })
      ).then(resp => {
        return {
          successes: resp.filter((item) => !item.code),
          errors: resp.filter((item) => item.code),
          response: resp,
        }
      })
    }

    return shared.tryCallback(promise, cb)
  }
}

module.exports = HsmSigner