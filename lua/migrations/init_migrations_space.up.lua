box.schema.create_space('_migrations_space_', { if_not_exists = true, format={
    {'id',type='string'},
    {'executed_at',type='string'},
}})

box.space._migrations_space_:create_index('id', {parts = {'id'}, if_not_exists = true, unique = true})