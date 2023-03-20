# haruhi

Welcome! This is Sandy's take on making a simple yet powerful enough
HTTP library. Inspired by [python's requests](https://github.com/psf/requests).

![haruhi](haruhi.gif)

## Motivation

When it comes to making network requests, I find myself writing a lot of repetitive
go code, you know the gist of getting a json response,

```go
req, err := http.NewRequestWithContext(context, method, URLpath, body)
if err != nil { ... }

params := req.URL.Query()
params.Add(key, value)
// more great parameters you need
req.URL.RawQuery = params.Encode()

resp, err := client.Do(req)
if err != nil { ... }
defer resp.Body.Close()

// do a response status check if needed...

ans := &SomeStruct{}
if err != json.NewDecoder(resp.Body).Decode(ans); err != nil { ... }

// whatever other magic you need to do
```

This is still a condensed version! The example above skips a couple of steps for brevity,
like setting timeouts, deadlines, proper reading, buffering, progress, etc. 

You can end up making something of your own, but repeating the same code has many issues,
such as forgetting to close the response body or set proper contexts with plenty more.

## Examples

What if instead, all you had to write was

```go
ans := &SomeStruct{}
err := haruhi.URL(URLpath).
    Method(method).
    Params(params).
    Body(body).
    Timeout(time.Minute).
    ResponseJSON(ans)
```

and it did everything for you? You can even simplify it with haruhi's sensible defaults, such as
running a `GET` request by default, etc, so you could even do

```go
respStr, err := haruhi.URL(url).Get()
```

Haruhi of course, supports more funtionality that is aimed to be simple and straight-forward.
Entire codebase is documented, so head on to [go docs](https://pkg.go.dev/github.com/thecsw/haruhi)
to see what else she can do. 

Have fun requesting!

## Similar projects

There is a golang [requests](https://github.com/carlmjohnson/requests) project by Carl M. Johnson,
who wrote a nice [blog post](https://blog.carlmjohnson.net/post/2021/requests-golang-http-client/) 
about needing a nice http library to avoid many of mistakes mentioned above. It has been around for
much longer (around 2 years?) and I've tried to use it. However, it may have been me misreading 
methods, or something else, but my requests weren't passing as I expected them to. And it's a powerful
library, which can do roundtrip logging, redirect checkers, etc. However, most of the time, I want
a dead-simple interface that "just works" and stays close to how I make fully proper go requests. 
Almost as if it's a bunch of macros.

I also found an older [requests library](https://github.com/asmcos/requests), which I've never used
but wanted to mention nonetheless.

## Why Haruhi?

Because [Endless Eight is a cinematic masterpiece](https://letterboxd.com/thecsw/film/the-melancholy-of-haruhi-suzumiya/) and [Haruhi is life itself](https://haruhi.fandom.com/wiki/Haruhi_Suzumiya).
