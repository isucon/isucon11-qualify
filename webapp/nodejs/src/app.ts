import fs from "fs";
import path from "path";

import session from "cookie-session";
import express from "express";
import jwt from "jsonwebtoken";
import morgan from "morgan";
import mysql from "mysql2/promise";

const dbinfo: mysql.PoolOptions = {
  host: process.env["MYSQL_HOSTNAME"] ?? "127.0.0.1",
  port: Number.parseInt(process.env["MYSQL_PORT"] ?? "3306"),
  user: process.env["MYSQL_USER"] ?? "isucon",
  database: process.env["MYSQL_DATABASE"] ?? "isucondition",
  password: process.env["MYSQL_PASS"] || "isucon",
  connectionLimit: 10,
};

const pool = mysql.createPool(dbinfo);

const app = express();

app.use(express.static("../public"));
app.use(morgan("combined"));
app.use(
  session({
    secret: process.env["SESSION_KEY"] ?? "isucondition",
    name: "isucondition",
    maxAge: 60 * 60 * 24 * 1000 * 30,
  })
);
app.set("cert", fs.readFileSync("../ec256-public.pem"));

"/ /condition /isu/:jia_isu_uuid /register /login"
  .split(" ")
  .forEach((frontendPath) => {
    app.get(frontendPath, (_req, res) => {
      res.sendFile(path.resolve("../public", "index.html"));
    });
  });

// POST /initialize
app.post("/initialize", async (_req, res) => {
  // TODO
  res.status(200).json({ language: "nodejs" });
});

// POST /api/auth
app.post("/api/auth", async (req, res) => {
  const db = await pool.getConnection();
  const authHeader = req.headers.authorization ?? "";
  const token = authHeader.startsWith("Bearer ")
    ? authHeader.slice(7)
    : authHeader;
  try {
    const decoded = jwt.verify(token, req.app.get("cert")) as jwt.JwtPayload;
    if (!("jia_user_id" in decoded)) {
      return res.status(400).send("invalid JWT payload");
    }
    const jiaUserId = decoded["jia_user_id"];
    if (typeof jiaUserId !== "string") {
      return res.status(400).send("invalid JWT payload");
    }
    await db.query(
      "INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)",
      jiaUserId
    );
    await db.commit();
    req.session = { jia_user_id: jiaUserId };
    res.status(200).send();
  } catch (err) {
    console.error(`jwt validation error: ${err}`);
    res.status(403).send("forbidden");
  } finally {
    db.release();
  }
});

// GET /api/user/me
app.get("/api/user/me", async (req, res) => {
  if (!(req.session && "jia_user_id" in req.session)) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }
  const jia_user_id = req.session["jia_user_id"];
  res.status(200).json({ jia_user_id });
});

// POST /api/signout
app.post("/api/signout", async (req, res) => {
  if (!(req.session && "jia_user_id" in req.session)) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }
  req.session = null;
  res.status(200).send();
});

app.listen(process.env.PORT ?? 3000, () => {});
