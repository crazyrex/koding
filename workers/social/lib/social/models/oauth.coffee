bongo    = require 'bongo'
{secure} = bongo

module.exports = class OAuth extends bongo.Base
  @share()

  @set
    sharedMethods   :
      static        : ['getUrl']

  @getUrl = secure (client, provider, callback)->
    switch provider
      when "github"
        {clientId} = KONFIG.github
        url = "https://github.com/login/oauth/authorize?client_id=#{clientId}&scope=user:email"
        callback null, url
      when "facebook"
        {clientId, redirectUri} = KONFIG.facebook
        url = "https://facebook.com/dialog/oauth?client_id=#{clientId}&redirect_uri=#{redirectUri}"
        callback null, url
      when "odesk"
        @getOdeskUrl (err, url, requestToken, requestTokenSecret)=>
          if err then callback err
          else
            @saveOdeskTokens client, url, requestToken, requestTokenSecret, (err)->
              callback err, url

  @getOdeskUrl = (callback)->
    Odesk         = require 'node-odesk'
    config        = KONFIG.odesk
    {key, secret} = config

    odesk = new Odesk key, secret
    odesk.OAuth.getAuthorizeUrl callback

  @saveOdeskTokens = (client, url, requestToken, requestTokenSecret, callback)->
    JSession = require './session'
    JSession.one {clientId: client.sessionToken}, (err, session) =>
      if err then callback err
      else
        odesk = {requestToken, requestTokenSecret}
        session.update $set: {"foreignAuth.odesk" : odesk}, (err)->
          callback err, url
