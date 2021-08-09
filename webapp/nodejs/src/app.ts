import fs from "fs";
import path from "path";

import session from "cookie-session";
import express from "express";
import jwt from "jsonwebtoken";
import morgan from "morgan";
import mysql, { FieldPacket, RowDataPacket, Connection } from "mysql2/promise";

declare global {
  namespace CookieSessionInterfaces {
    interface CookieSessionObject {
      jia_user_id?: string;
    }
  }
}

interface IsuRow extends RowDataPacket {
  id: number;
  jia_isu_uuid: string;
  name: string;
  image: Buffer;
  character: string;
  jia_user_id: string;
  created_at: Date;
  updated_at: Date;
}

interface GetIsuListResponse {
  id: number;
  jia_isu_uuid: string;
  name: string;
  character: string;
  latest_isu_condition?: GetIsuConditionResponse;
}

interface GetIsuConditionResponse {
  jia_isu_uuid: string;
  isu_name: string;
  timestamp: number;
  is_sitting: boolean;
  condition: string;
  condition_level: string;
  message: string;
}

interface IsuConditionRow extends RowDataPacket {
  id: number;
  jia_isu_uuid: string;
  timestamp: Date;
  is_sitting: boolean;
  condition: string;
  message: string;
  created_at: Date;
}

const sessionName = "isucondition";
// const conditionLimit = 20;
const frontendContentsPath = "../public";
const jwtVerificationKeyPath = "../ec256-public.pem";
// defaultIconFilePath       = "../NoImage.png"
// defaultJIAServiceURL      = "http://localhost:5000"
// mysqlErrNumDuplicateEntry = 1062
const conditionLevelInfo = "info";
const conditionLevelWarning = "warning";
const conditionLevelCritical = "critical";
// const scoreConditionLevelInfo     = 3
// const scoreConditionLevelWarning  = 2
// const scoreConditionLevelCritical = 1

const dbinfo: mysql.PoolOptions = {
  host: process.env["MYSQL_HOSTNAME"] ?? "127.0.0.1",
  port: Number.parseInt(process.env["MYSQL_PORT"] ?? "3306"),
  user: process.env["MYSQL_USER"] ?? "isucon",
  database: process.env["MYSQL_DATABASE"] ?? "isucondition",
  password: process.env["MYSQL_PASS"] || "isucon",
  connectionLimit: 10,
  timezone: "local",
};

const pool = mysql.createPool(dbinfo);

const app = express();

app.use("/assets", express.static(frontendContentsPath + "/assets"));
app.use(morgan("combined"));
app.use(
  session({
    secret: process.env["SESSION_KEY"] ?? "isucondition",
    name: sessionName,
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

// POST /api/signout
app.post("/api/signout", async (req, res) => {
  const jia_user_id = req.session?.jia_user_id;
  if (!jia_user_id) {
    console.error("you are not signed in");
    return res.status(401).type("text").send("you are not signed in");
  }
  req.session = null;
  res.status(200).send();
});

// GET /api/user/me
app.get("/api/user/me", async (req, res) => {
  const jia_user_id = req.session?.jia_user_id;
  if (!jia_user_id) {
    console.error("you are not signed in");
    return res.status(401).type("text").send("you are not signed in");
  }
  res.status(200).json({ jia_user_id });
});

// GET /api/isu
app.get("/api/isu", async (req, res) => {
  const jiaUserId = req.session?.jia_user_id;
  if (!jiaUserId) {
    console.error("you are not signed in");
    return res.status(401).send("you are not signed in");
  }

  const db = await pool.getConnection();
  try {
    await db.beginTransaction();
    const [isuList]: [IsuRow[], FieldPacket[]] = await db.query(
      "SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
      [jiaUserId]
    );
    const responseList: Array<GetIsuListResponse> = [];
    for (const isu of isuList) {
      let foundLastCondition = true;
      const [[lastCondition, ..._]]: [IsuConditionRow[], FieldPacket[]] =
        await db.query(
          "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1",
          [isu.jia_isu_uuid]
        );
      if (lastCondition.length) {
        foundLastCondition = false;
      }
      let formattedCondition = undefined;
      if (foundLastCondition) {
        const [conditionLevel, err] = calculateConditionLevel(
          lastCondition.condition
        );
        if (err) {
          console.error(`failed to get condition level: ${err}`);
          return res.status(500).send();
        }
        formattedCondition = {
          jia_isu_uuid: lastCondition.jia_isu_uuid,
          isu_name: isu.name,
          timestamp: lastCondition.timestamp.getTime() / 1000,
          is_sitting: lastCondition.is_sitting,
          condition: lastCondition.condition,
          condition_level: conditionLevel,
          message: lastCondition.message,
        };
      }
      responseList.push({
        id: isu.id,
        jia_isu_uuid: isu.jia_isu_uuid,
        name: isu.name,
        character: isu.character,
        latest_isu_condition: formattedCondition,
      });
    }
    await db.commit();
    res.status(200).json(responseList);
  } catch (err) {
    console.error(`db error: ${err}`);
    return res.status(500).send();
  } finally {
    db.release();
  }
});

/*
interface GetAllIsuConditionsQuery {
  cursor_end_time: string;
  cursor_jia_isu_uuid: string;
  condition_level: string;
  start_time?: string;
  limit?: string;
}

// GET /api/condition
app.get(
  "/api/condition",
  async (req: express.Request<{}, any, any, GetAllIsuConditionsQuery>, res) => {
    const jia_user_id = req.session?.jia_user_id;
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

    // optional query param
    let startTime = new Date().getTime() / 1000;
    if (req.query.start_time) {
      startTime = parseInt(req.query.start_time, 10);
      if (Number.isNaN(startTime)) {
        console.error("bad format: start_time");
        return res.status(400).send("bad format: start_time");
      }
    }
    const limit = req.query.limit
      ? parseInt(req.query.limit, 10)
      : conditionLimit;
    if (Number.isNaN(limit)) {
      console.error("bad format: limit");
      return res.status(400).send("bad format: limit");
    }

    const db = await pool.getConnection();
    try {
      const [isuList]: [IsuRow[], FieldPacket[]] = await db.query(
        "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `is_deleted` = false",
        [jia_user_id]
      );
      if (isuList.length === 0) {
        return res.status(200).json(isuList);
      }

      // TODO:
      const conditionsResponse = await Promise.all(
        isuList.map(async (isu) => {
          const condition = await getIsuConditionsFromDB(
            isu.jia_isu_uuid,
            cursorEndTime + 1,
            conditionLevel,
            startTime,
            limit + 1,
            isu.name
          );
          return condition;
        })
      );

      res.status(200).json(isuList);
    } catch (err) {
      console.error(`db error: ${err}`);
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

const getIsuConditionsFromDB = async (
  jiaIsuUUID: string,
  cursorEndTime: number,
  conditionLevel: Set<string>,
  startTime: number,
  limit: number,
  isuName: string
) => {
  // TODO:
};
*/

// conditionのcsvからcondition levelを計算
function calculateConditionLevel(condition: string): [string, Error?] {
  var conditionLevel: string;

  const warnCount = Array.from(condition.matchAll(/=true/g)).length;
  switch (warnCount) {
    case 0:
      conditionLevel = conditionLevelInfo;
      break;
    case 1:
    case 2:
      conditionLevel = conditionLevelWarning;
      break;
    case 3:
      conditionLevel = conditionLevelCritical;
      break;
    default:
      return ["", new Error("unexpected warn count")];
  }
  return [conditionLevel, undefined];
}

app.listen(process.env.PORT ?? 3000, () => {});
