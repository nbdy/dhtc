from maga import Maga, proper_infohash
from mala import get_metadata
from time import time
from loguru import logger as log

from dhtc import Database


class Crawler(Maga):
    used_hashes = []

    def __init__(self, db: Database, bootstrap_nodes):
        Maga.__init__(self, bootstrap_nodes=bootstrap_nodes, interval=0.1)
        self.db = db

    async def handler(self, info_hash, addr):
        log.debug("{} > {}", self.addr_str(addr), info_hash)

    @staticmethod
    def addr_str(addr):
        return "%s:%i" % (addr[0], addr[1])

    async def handle_announce_peer(self, info_hash, addr, peer_addr):
        log.debug("{} - {} > {}", self.addr_str(addr), self.addr_str(peer_addr), info_hash)
        if info_hash in self.used_hashes:
            return  # we are currently pulling data for this one
        e = self.db.get(info_hash, addr, peer_addr)
        if e.seen > 0:
            if e.addr != addr:
                e.prev_addrs.append(e.addr)
                e.addr = addr
            if e.peer_addr != peer_addr:
                e.prev_peer_addrs.append(e.peer_addr)
                e.peer_addr = peer_addr
            e.last_seen_list.append(e.last_seen)
        e.seen += 1
        e.last_seen = time()

        self.used_hashes.append(info_hash)
        meta_info = await get_metadata(info_hash, peer_addr[0], peer_addr[1], loop=self.loop)
        if meta_info:
            try:
                meta_info["proper_infohash"] = proper_infohash(info_hash)
                print(info_hash, meta_info["proper_infohash"])
            except Exception as e:
                print(e)
                pass
            if not isinstance(meta_info, dict):
                log.debug("{} announced malformed data '{}'", addr, meta_info)
            else:
                e.meta_infos.append(meta_info)

        self.used_hashes.remove(info_hash)
        self.db.save(e)
        log.info(e.__dict__)
