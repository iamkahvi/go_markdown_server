// server --X--> client
type ServerMessage = {
  status: "leader" | "follower";
  data: string; // either initial state or live updates from the file
};

// client --X--> server
type ClientMessage = {
  data: string; // live updates from the editor
};

// Case 1
// 1.
// client --X--> server
{
  type: "first"
}
// 2.
// server --X--> client
{
  status: "leader",
  data: "I'm the initial state of the file"
}
// 3.
// client --X--> server
{
  data: "I'm the new state of the editor"
}

// Case 2
// 1.
// client --X--> server
{
  type: "first"
}
// 2.
// server --X--> client
{
  status: "follower",
  data: "I'm the initial state of the file"
}
// 3.
// server --X--> client
{
  status: "follower",
  data: "I'm the new initial state of the file"
}

// Case 3
// 1.
// client --X--> server
{
  type: "first"
}
// 2.
// server --X--> client
{
  status: "follower",
  data: "I'm the initial state of the file"
}
// 3.
// server --X--> client
{
  status: "leader",
  data: "I'm the new initial state of the file"
}
// 4.
// client --X--> server
{
  data: "I'm the new state of the editor"
}
