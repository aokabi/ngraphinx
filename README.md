# ngraphinx
nginxのaccess.logを時系列のグラフにするcliツール


![](./docs/sample.png)

## install
`go install github.com/aokabi/ngraphinx@latest`

## usage
`ngraphinx --path="./sample-accesslog/access.log" --aggregates="/isu/*,/api/condition/*,/?jwt"`

## contribution
welcome!
