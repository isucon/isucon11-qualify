import fs from "fs";
import path from "path";

import session from "cookie-session";
import express from "express";
import jwt from "jsonwebtoken";
import morgan from "morgan";
import mysql from "mysql2/promise";

interface GetIsuConditionResponse {
  jia_isu_uuid: string;
  isu_name: string;
  timestamp: number;
  is_sitting: boolean;
  condition: string;
  condition_level: string;
  message: string;
}

const sessionName = "isucondition";
// conditionLimit            = 20
const isuListLimit = 200; // TODO 修正が必要なら変更
const frontendContentsPath = "../public";
const jwtVerificationKeyPath = "../ec256-public.pem";
// defaultIconFilePath       = "../NoImage.png"
// defaultJIAServiceURL      = "http://localhost:5000"
// mysqlErrNumDuplicateEntry = 1062
// )

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

app.use(express.static(frontendContentsPath));
app.use(morgan("combined"));
app.use(
  session({
    secret: process.env["SESSION_KEY"] ?? "isucondition",
    name: "isucondition",
    maxAge: 60 * 60 * 24 * 1000 * 30,
  })
);
app.set("cert", fs.readFileSync(jwtVerificationKeyPath));

["/", "/condition", "/isu/:jia_isu_uuid", "/register", "/login"].forEach(
  (frontendPath) => {
    app.get(frontendPath, (_req, res) => {
      res.sendFile(path.resolve("../public", "index.html"));
    });
  }
);

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
  const jia_user_id = req.session?.["jia_user_id"];
  if (!jia_user_id) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }
  res.status(200).json({ jia_user_id });
});

// POST /api/signout
app.post("/api/signout", async (req, res) => {
  const jia_user_id = req.session?.["jia_user_id"];
  if (!jia_user_id) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }
  req.session = null;
  res.status(200).send();
});

// GET /api/isu
app.get("/api/isu", async (req, res) => {
  const jia_user_id = req.session?.["jia_user_id"];
  if (!jia_user_id) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }

  const limit =
    typeof req.query["limit"] === "string"
      ? parseInt(req.query["limit"], 10)
      : isuListLimit;
  if (Number.isNaN(limit)) {
    console.error("invalid value: limit");
    return res.status(400).send("invalid value: limit");
  }

  const db = await pool.getConnection();
  try {
    const [isuList] = await db.query(
      "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false ORDER BY `created_at` DESC LIMIT ?",
      [jia_user_id, limit]
    );
    res.status(200).json(isuList);
  } catch (err) {
    console.error(`db error: ${err}`);
    return res.status(500).send();
  } finally {
    db.release();
  }
});

interface GetAllIsuConditionsQuery {
  cursor_end_time: string;
  cursor_jia_isu_uuid: string;
  condition_level: string;
}

// GET /api/condition
app.get(
  "/api/condition",
  async (req: express.Request<{}, any, any, GetAllIsuConditionsQuery>, res) => {
    const jia_user_id = req.session?.["jia_user_id"];
    if (!jia_user_id) {
      console.error("you are not signed in");
      return res.status(401).send("you are not signed in");
    }

    // required query params
    const cursorEndTime = parseInt(req.query.cursor_end_time, 10);
    if (Number.isNaN(cursorEndTime)) {
      console.error("bad format: cursor_end_time");
      return res.status(400).send("bad format: cursor_end_time");
    }
    const cursorJIAIsuUUID = req.query.cursor_jia_isu_uuid;
    if (!cursorJIAIsuUUID) {
      console.error("cursor_jia_isu_uuid is missing");
      return res.status(400).send("cursor_jia_isu_uuid is missing");
    }
    const conditionLevelCSV = req.query.condition_level;
    if (!conditionLevelCSV) {
      console.error("condition_level is missing");
      return res.status(400).send("condition_level is missing");
    }
    const conditionLevel = new Set(conditionLevelCSV.split(","));

    // TODO:

    res.status(200).json([]);
  }
);

app.listen(process.env.PORT ?? 3000, () => {});
