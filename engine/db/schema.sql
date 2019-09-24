create table engine_cursors (
   id varchar(255) not null,
   `cursor` bigint not null,
   updated_at datetime(3) not null,

   primary key (id)
);

create table engine_events (
  id bigint not null auto_increment,
  foreign_id bigint not null,
  timestamp datetime(3) not null,
  type int not null,

  primary key (id)
);

create table engine_matches (
  id bigint not null auto_increment,
  status int not null,
  team varchar(255) not null,
  players int not null,
  summary text,

  created_at datetime(3) not null,
  updated_at datetime(3) not null,


  primary key (id),
  unique by_team_status (team,status)
);

create table engine_rounds (
  id bigint not null auto_increment,
  match_id bigint not null,
  `index` int not null,
  status int not null,
  team varchar(255) not null,
  state text,
  error text,

  created_at datetime(3) not null,
  updated_at datetime(3) not null,

  primary key (id),
  unique by_team_status (team,`index`)
);
