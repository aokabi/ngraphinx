# ngraphinx
nginxのaccess.logを時系列のグラフにするcliツール

## install
`go install github.com/aokabi/ngraphinx@latest`

## usage
`ngraphinx --path="./sample-accesslog/access.log" --aggregates="/isu/*,/api/condition/*,/?jwt"`
