# SPEAR - Simple, P2P, Encrypted And Real-Time

Spear is a VoIP program that doesn’t require any central server. It uses a CUI interface so you can run it on a terminal or even tty if you want. Packets are encrypted using ChaCha20Poly1305 and shared secrets are generated using a pre-shared public key from a peer and the user’s own secret, specified in a config file.

Example config.conf:

```
[Client]
sk = g7suVU4IhGd8slx5q618dz0NBgMujeWSu1r2eKHJBSg= #user’s secret key
candidates = 0.0.0.0:3412, 192.168.0.1:54361 #ip:port
#spear will try to bind to one of the ‘candidates’

[Peer]
pk = D4VwZ+mrsWV8yyQSlty7F82HNDpNDM5AzJV1VAMC2jc= #peer’s public key
candidates = 123.123.123.123:62162, 321.321.321.231:41231
#a peer can has multiple possible endpoints

[Peer]
pk = KutbwzJ0d1mfrijI8r0+lfPQLdbIsa0UV7QvuTF5QXY=
candidates = 321.123.123.312:12321
name = Friend 1 #optional
```
# TODO:
Screen sharing


