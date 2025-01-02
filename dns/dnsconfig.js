// read more about the dnscontrol tool here: <https://docs.dnscontrol.org/>

var CF_MAX_TTL = TTL('1d')

D('indocker.app', NewRegistrar('none'), DnsProvider(NewDnsProvider('cloudflare')),
  // all subdomains
  A('*', '127.0.0.1', CF_MAX_TTL),
  AAAA('*', '::1', CF_MAX_TTL),

  // special case for the docker (https://habr.com/en/post/714916/#comment_25196630)
  A('*.x-docker', '172.17.0.1'),

  // an alias for the domain with certs (https://github.com/tarampampam/indocker-app/issues/79)
  CNAME('x-cert', 'indocker-app-certs.pages.dev.', CF_PROXY_ON),

  // index page
  ALIAS('@', 'indocker-app-index.pages.dev.', CF_PROXY_ON), // aka CNAME for the CF

  // disallow emails
  TXT('_dmarc', 'v=DMARC1; p=reject; sp=reject; adkim=s; aspf=s;'),
  TXT('*._domainkey', 'v=DKIM1; p='),
  TXT('@', 'v=spf1 -all')
)
