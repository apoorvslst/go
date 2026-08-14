# 🏦 Lecture 9: Database Transaction Isolation Levels (MySQL vs Postgres)

---

## 🦅 High-Level Overview (Bird's Eye View)

Pichle lecture mein humne dekha ki ACID properties mein **"I" (Isolation)** ka kya matlab hota hai aur Deadlocks ko kaise roka jaye. Aaj ka lecture bilkul **Theory + Practical DB Console** ka mix hai jisme hum kisi naye Go code ko nahi likhenge, balki direct database engines (MySQL aur Postgres) se raw SQL run karke baat karenge.

Jab 2 transactions ek saath chal rahi hon (concurrently), toh ek doosre ke raste mein aane ki wajah se kuch "Read Phenomena" (Data dikhne mein gadbadi) ho sakti hain:
1. **Dirty Read:** Ek transaction ne abhi data modify kiya par commit nahi kiya, aur dusre ne wo "kacha" data padh liya. Agar pichla transaction fail ho jaye (rollback), toh dusre transaction ke paas galat data aagaya.
2. **Non-repeatable Read:** Ek transaction do baar same record padhti hai, lekin dono baar result alag aata hai (kyunki beech mein kisi aur transaction ne aakar data update aur commit kar diya).
3. **Phantom Read:** Tumne do baar same "Search condition" (jaise balance > 50) chalayi, lekin doosri baar mein tumhe result mein alag set of rows mili (kyunki kisi ne naya record insert ya delete kar diya tha).
4. **Serialization Anomaly:** Jab aisi state create ho jaye jo possible hi nahi hoti agar wo saari transactions ek ke baad ek (serially) chali hoti.

In problems ko rokne ke liye **4 Isolation Levels** banaye gaye hain:
- **Read Uncommitted:** Sabse lowest. Dirty read bhi allow karta hai.
- **Read Committed:** Dirty read rokta hai. (Postgres ka Default).
- **Repeatable Read:** Dirty aur Non-repeatable read dono ko rokta hai. (MySQL ka Default).
- **Serializable:** Sabse highest level. Sab gadbadiyaan rokta hai, par isme locks zyada lagte hain aur deadlocks common hote hain.

Is lecture mein in sabko practically manually MySQL aur Postgres terminal mein chala kar test kiya gaya hai.

---

## 🗺️ Step-by-Step: Kya Padhaya Gaya (The DB Testing)?

### Step 1: Read Uncommitted Level & Dirty Reads
- MySQL ke 2 consoles open kiye aur isolation level `Read Uncommitted` set kiya.
- **Tx 1** ne balance 100 se 90 kiya (commit **nahi** kiya).
- **Tx 2** ne usko read kiya aur dekha ki balance `90` hai. Ye **Dirty Read** hai.
- *Fact:* Postgres mein `Read Uncommitted` practically `Read Committed` jaisa hi kaam karta hai. Ye lowest level support hi nahi karta kyunki ye bohot dangerous hai.

### Step 2: Read Committed Level (Default in Postgres)
- Isolation level badhakar `Read Committed` kiya.
- Tx 1 ne update maar ke commit **nahi** kiya, toh Tx 2 ko purana value (`100`) hi dikha (Dirty read rukk gaya!).
- Jaise hi Tx 1 ne **commit** kiya, tab jake Tx 2 ko naya data (`80`) dikh gaya. Yani ek hi transaction ke andar 2 baar read karne pe alag data dikha — Ye hai **Non-repeatable read**.

### Step 3: Repeatable Read Level (Default in MySQL)
- Level kiya `Repeatable Read`.
- Tx 1 ne data update aur commit kar diya, phir bhi Tx 2 ko wahi pehli wali value hi dikh rahi thi. Isne **Non-repeatable read** ko block kar diya.
- Lekin ek twist hai: Agar Tx 2 us purani dikh rahi value ke upar update query chalaye, toh MySQL naye number ke hisab se math thodi ajeeb kar deta hai (Update chal jata hai). Postgres yahan par error de deta hai `could not serialize access due to concurrent update`, jo ki zyada badiya aur safe logic hai.

### Step 4: Serializable Level (The Boss Level)
- Sabse strict level. Yahan **Serialization Anomaly** test ki gayi (Do log ek saath poore table ka total `Sum` nikalte hain aur naye record mein add karna chahte hain).
- Lower levels pe 2 identical record ban jate, jabki serial chalta to alag alag sum aata.
- **MySQL mein:** Serializable level automatically SELECT queries ko `SELECT FOR SHARE` (lock) bana deta hai, isliye doosra update nahi kar pata (Lock wait timeout ya deadlock lag jata hai).
- **Postgres mein:** Postgres locks pe depend nahi karta, iska apna intelligent "dependency checking mechanism" hai. Ye turant doosre transaction par error maar deta hai: `could not serialize access due to read/write dependencies`.

---

## 🛠 Main Takeaway (As a Developer kya dhyan rakhna hai?)

1. MySQL mein default isolation `Repeatable Read` hai, lekin Postgres mein `Read Committed` hai.
2. Jab tum database ko completely strict (`Serializable`) karte ho, tab consistency ekdum perfect hoti hai lekin **Errors, Timeouts aur Deadlocks** ke chances kaafi badh jaate hain.
3. **Golden Rule:** Jab bhi tum high isolation level DB me code likh rahe ho, hamesha apne backend logic (Golang/Python) mein ek **"Retry Mechanism"** implement karo. Agar aisi temporary dependency wali error aaye, toh code khudse us transaction ko wapas restart karke successfully commit kar le.

---
*(Note: As this lecture is purely theoretical and uses raw SQL queries to explain database behaviors directly on the DB engines, there is no new Golang code written in this lecture. We have just carried over the Golang code from lecture 8 into this directory so that the codebase remains complete and consistent with the project's progress).*
