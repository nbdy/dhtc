import shelve
from random import choice

from .DBEntry import DBEntry


class Database(object):
    def __init__(self, path="dht"):
        self.db = shelve.open(path)

    def has_key(self, key):
        return key in self.db.keys()

    def get(self, key, addr=None, peer_addr=None) -> DBEntry:
        return self.db.get(key, DBEntry(key, addr, peer_addr))

    def get_random_key(self):
        k = list(self.db.keys())
        if len(k) > 0:
            return choice(k)
        return None

    def get_random_entry(self):
        k = self.get_random_key()
        if k:
            return self.get(k)
        return None

    def get_random_title(self):
        e = self.get_random_entry()
        if e and len(e.meta_infos) > 0:
            try:
                return choice(e.meta_infos)["title"]
            except KeyError:
                pass
        return "no entries yet"

    def get_count(self):
        return len(self.db.keys())

    def save(self, e: DBEntry):
        self.db[e.info_hash] = e
        self.db.sync()

    def close(self):
        self.db.close()

    def get_x_random(self, x=50, _t=3):
        r = []
        q = 0
        for _ in range(x):
            if q == _t:
                break
            e = self.get_random_entry()
            if e is None:
                break
            if e not in r:
                r.append(e)
            else:
                q += 1
        return r
