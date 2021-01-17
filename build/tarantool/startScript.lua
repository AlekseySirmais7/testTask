#!/usr/bin/env tarantool

box.cfg{
}

box.schema.user.grant('guest', 'read,write,execute', 'universe')
box.schema.user.passwd('admin', 'admin')

---------------------------------------

s = box.schema.space.create('statuses', { if_not_exists = false, engine = 'memtx' })

s:format({
                { name = 'id', type = 'integer' },
                { name = 'status', type = 'string' },
            })


s:create_index('primary', { type = 'tree', parts = { 'id' } })

-----------------------------------
