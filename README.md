# ngraphinx
nginxのaccess.logを時系列のグラフにするcliツール


![](./docs/sample.png)

## install
`go install github.com/aokabi/ngraphinx/v2@latest`

## usage
### image
- `ngraphinx image --path="./sample-accesslog/access.log" --aggregates="/isu/*,/api/condition/*,/?jwt"`
- open `{2006-01-02T15:04:05Z07:00}.png`

### chartjs
- `ngraphinx chartjs --path="./sample-accesslog/access.log" --aggregates="/isu/*,/api/condition/*,/?jwt"`
- open `{2006-01-02T15:04:05Z07:00}.html`


## contribution
welcome!
