# RubiksKube
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
Easy kubernetes for ubuntu bare metal

Install RubiksKube on all servers
```sh
$ go get -u github.com/JonathanHeinz/RubiksKube
```
Init master
```sh
$ rubikskube
```
Init a node
```sh
$ rubikskube -add-node={HASH}
```
Now just have fun with your small kubernetes cluster :grimacing: :thumbsup:

##### ToDo
- [ ] Integrate Cobra for better CLI feeling
- [ ] Add docs
- [ ] Add tests
- [ ] Add better Docker interaction
