# 0012: JWTの署名アルゴリズム検証をWithValidMethodsで宣言的に行う

## Status

Accepted

## Context

`internal/auth.ParseToken`は、`keyFunc`内で`t.Method.(*jwt.SigningMethodHMAC)`という型アサーションによって署名アルゴリズムがHMAC系であることを確認していた([ADR-0004](./0004-jwt-auth.md)のアルゴリズム混同攻撃対策)。動作としては正しいが、`golang-jwt/jwt/v5`にはこのチェックを`ParserOption`として宣言的に書ける`jwt.WithValidMethods`が用意されている。Step3提出後のレビューで、可読性の観点から指摘を受けた。

## Decision

`keyFunc`内の型アサーションを削除し、`jwt.ParseWithClaims`の呼び出しに`jwt.WithValidMethods([]string{"HS256"})`を追加する。このオプションは`keyFunc`が呼ばれるより前に、ライブラリ内部でトークンヘッダーの`alg`を検証する。

## Consequences

- 「HS256以外は受け付けない」という意図がコード上に明示され、`keyFunc`は鍵を返すことだけに専念できる
- セキュリティ上の効果(alg=none攻撃・アルゴリズム混同攻撃への耐性)は変更前と同じ
