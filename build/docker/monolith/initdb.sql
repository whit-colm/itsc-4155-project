--- YES this is probably not great security practice but hear me out:
--- 1. The monolith builds are not supposed to be used in production
--- 2. Even if some bone-head *does* use it in prod, the db is *NEVER*
---    exposed to the network
---
--- So that's why I am comfortable putting 
CREATE DATABASE jaws;
CREATE USER jaws WITH ENCRYPTED PASSWORD 'reallybadunsafesecuritypractice'
GRANT ALL PRIVLIGES ON DATABASE jaws TO jaws;