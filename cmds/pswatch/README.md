# pswatch
For android only

## Build

Test

    ./build.sh

Production

    sh gox_build.sh

## HTTP接口

启动 `./pswatch -l -port 16118

```
POST localhost:16118/api/v1/perf
```

参数

名字 | 类型   | 例子
-----|--------|-----------
name | string | "myfps"
data | float  | 15.5
