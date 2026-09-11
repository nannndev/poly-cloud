# 07 — Sequence Flows

## 1. Connect Account (OAuth)
```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant BE as Backend
    participant EN as Engine (rclone)
    participant P as Provider
    participant DB as Postgres

    U->>FE: klik "Tambah Account" (pilih provider)
    FE->>BE: POST /accounts/connect {provider,label}
    BE->>EN: minta auth URL (OAuth)
    EN-->>BE: auth_url
    BE-->>FE: {auth_url}
    FE->>P: redirect ke consent
    U->>P: login + consent
    P-->>FE: redirect callback ?code
    FE->>BE: POST /accounts/callback {code,state}
    BE->>EN: tukar code -> token
    EN->>P: exchange
    P-->>EN: access+refresh token
    EN-->>BE: token
    BE->>DB: simpan account + token (AES-GCM)
    BE-->>FE: {account_id, status:active}
    BE->>EN: initial sync (List)
    EN->>P: list files
    P-->>EN: metadata
    EN-->>BE: entries
    BE->>DB: tulis files_index + file_blocks
```

## 2. Upload + Smart Routing (Model A)
```mermaid
sequenceDiagram
    participant FE as Frontend
    participant BE as Backend
    participant R as Router
    participant DB as Postgres
    participant EN as Engine
    participant P as Provider

    FE->>BE: POST /files/upload (stream, size)
    BE->>DB: ambil accounts + kuota
    DB-->>BE: accounts[]
    BE->>R: Pick(accounts, size)
    alt ada account muat
        R-->>BE: account terpilih (most-free)
        BE->>EN: UploadStream(reader, remote, dest)
        EN->>P: rcat stream
        P-->>EN: ok (provider_ref)
        EN-->>BE: sukses
        BE->>DB: insert files_index + file_blocks(seq=0)
        BE-->>FE: {file_id, account_label, routed_by}
    else tak ada yang muat
        R-->>BE: ErrNoRoom
        BE-->>FE: 409 NO_ROOM (Model B akan chunk)
    end
```

## 3. Download (Stream-Through)
```mermaid
sequenceDiagram
    participant FE as Frontend
    participant BE as Backend
    participant DB as Postgres
    participant EN as Engine
    participant P as Provider

    FE->>BE: GET /files/{id}/download
    BE->>DB: cari file_blocks(file_id)
    DB-->>BE: block(account, provider_ref)
    BE->>EN: Download(remote, path, writer)
    EN->>P: cat stream
    P-->>EN: bytes
    EN-->>BE: pipe stream
    BE-->>FE: stream response (tanpa simpan di server)
```

## 4. Token Refresh (Lifecycle)
```mermaid
sequenceDiagram
    participant BE as Backend
    participant DB as Postgres
    participant EN as Engine
    participant P as Provider

    Note over BE: sebelum operasi apa pun
    BE->>DB: baca token + expires_at
    alt token masih valid
        BE->>EN: lanjut operasi
    else expired
        BE->>EN: refresh(refresh_token)
        EN->>P: request token baru
        alt refresh sukses
            P-->>EN: token baru
            EN-->>BE: token
            BE->>DB: simpan ulang (AES-GCM)
            BE->>EN: lanjut operasi
        else refresh gagal
            P-->>EN: error
            EN-->>BE: gagal
            BE->>DB: set status=needs_reconnect
            Note over BE: UI minta user re-auth
        end
    end
```

## 5. (Future) Upgrade File A → B
```mermaid
sequenceDiagram
    participant BE as Backend
    participant WS as WholeFileStore
    participant SP as Splitter
    participant R as Router
    participant EN as Engine
    participant DB as Postgres

    BE->>WS: baca file utuh (Model A)
    WS-->>BE: stream
    BE->>SP: pecah + (encrypt) -> N chunk
    loop tiap chunk
        BE->>R: Pick(accounts, chunkSize)
        R-->>BE: account
        BE->>EN: UploadStream(chunk)
    end
    BE->>DB: tulis N file_blocks(seq=0..N-1)
    BE->>DB: set files_index.is_chunked=true
    Note over BE: hapus objek lama; reversible
```
