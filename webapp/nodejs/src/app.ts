import { spawn } from "child_process";
import { readFileSync } from "fs";
import { readFile } from "fs/promises";
import path from "path";

import axios from "axios";
import session from "cookie-session";
import express from "express";
import jwt from "jsonwebtoken";
import morgan from "morgan";
import multer, { MulterError } from "multer";
import mysql, { RowDataPacket } from "mysql2/promise";
import qs from "qs";

interface Config extends RowDataPacket {
  name: string;
  url: string;
}

interface IsuResponse {
  id: number;
  jia_isu_uuid: string;
  name: string;
  character: string;
}

interface Isu extends IsuResponse, RowDataPacket {
  image: Buffer;
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

interface IsuCondition extends RowDataPacket {
  id: number;
  jia_isu_uuid: string;
  timestamp: Date;
  is_sitting: number;
  condition: string;
  message: string;
  created_at: Date;
}

interface InitializeResponse {
  language: string;
}

interface GetMeResponse {
  jia_user_id: string;
}

interface GraphResponse {
  start_at: number;
  end_at: number;
  data?: GraphDataPoint;
  condition_timestamps: number[];
}

interface GraphDataPoint {
  score: number;
  percentage: ConditionsPercentage;
}

interface ConditionsPercentage {
  sitting: number;
  is_broken: number;
  is_dirty: number;
  is_overweight: number;
}

interface GraphDataPointWithInfo {
  jiaIsuUUID: string;
  startAt: Date;
  data: GraphDataPoint;
  conditionTimeStamps: number[];
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

interface TrendResponse {
  character: string;
  info: TrendCondition[];
  warning: TrendCondition[];
  critical: TrendCondition[];
}

interface TrendCondition {
  isu_id: number;
  timestamp: number;
}

const sessionName = "isucondition_nodejs";
const conditionLimit = 20;
const frontendContentsPath = "../public";
const jiaJWTSigningKeyPath = "../ec256-public.pem";
const defaultIconFilePath = "../NoImage.jpg";
const defaultJIAServiceUrl = "http://localhost:5000";
const mysqlErrNumDuplicateEntry = 1062;
const conditionLevelInfo = "info";
const conditionLevelWarning = "warning";
const conditionLevelCritical = "critical";
const scoreConditionLevelInfo = 3;
const scoreConditionLevelWarning = 2;
const scoreConditionLevelCritical = 1;

if (!("POST_ISUCONDITION_TARGET_BASE_URL" in process.env)) {
  console.error("missing: POST_ISUCONDITION_TARGET_BASE_URL");
  process.exit(1);
}
const postIsuConditionTargetBaseURL =
  process.env["POST_ISUCONDITION_TARGET_BASE_URL"];
const dbinfo: mysql.PoolOptions = {
  host: process.env["MYSQL_HOST"] ?? "127.0.0.1",
  port: parseInt(process.env["MYSQL_PORT"] ?? "3306", 10),
  user: process.env["MYSQL_USER"] ?? "isucon",
  database: process.env["MYSQL_DBNAME"] ?? "isucondition",
  password: process.env["MYSQL_PASS"] || "isucon",
  connectionLimit: 10,
  timezone: "+09:00",
};
const pool = mysql.createPool(dbinfo);
const upload = multer();

const app = express();

app.use(morgan("combined"));
app.use("/assets", express.static(frontendContentsPath + "/assets"));
app.use(express.json());
app.use(
  session({
    secret: process.env["SESSION_KEY"] ?? "isucondition",
    name: sessionName,
    maxAge: 60 * 60 * 24 * 1000 * 30,
  })
);
app.set("cert", readFileSync(jiaJWTSigningKeyPath));
app.set("etag", false);

class ErrorWithStatus extends Error {
  public status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = new.target.name;
    this.status = status;
  }
}

async function getUserIdFromSession(
  req: express.Request,
  db: mysql.Connection
): Promise<string> {
  if (!req.session) {
    throw new ErrorWithStatus(500, "failed to get session");
  }
  const jiaUserId = req.session["jia_user_id"];
  if (!jiaUserId) {
    throw new ErrorWithStatus(401, "no session");
  }

  let cnt: number;
  try {
    [[{ cnt }]] = await db.query<(RowDataPacket & { cnt: number })[]>(
      "SELECT COUNT(*) AS `cnt` FROM `user` WHERE `jia_user_id` = ?",
      [jiaUserId]
    );
  } catch (err) {
    throw new ErrorWithStatus(500, `db error: ${err}`);
  }
  if (cnt === 0) {
    throw new ErrorWithStatus(401, "not found: user");
  }
  return jiaUserId;
}

async function getJIAServiceUrl(db: mysql.Connection): Promise<string> {
  const [[config]] = await db.query<Config[]>(
    "SELECT * FROM `isu_association_config` WHERE `name` = ?",
    ["jia_service_url"]
  );
  if (!config) {
    return defaultJIAServiceUrl;
  }
  return config.url;
}

interface PostInitializeRequest {
  jia_service_url: string;
}

function isValidPostInitializeRequest(
  body: PostInitializeRequest
): body is PostInitializeRequest {
  return typeof body === "object" && typeof body.jia_service_url === "string";
}

// POST /initialize
// サービスを初期化
app.post(
  "/initialize",
  async (
    req: express.Request<Record<string, never>, unknown, PostInitializeRequest>,
    res
  ) => {
    const request = req.body;
    if (!isValidPostInitializeRequest(request)) {
      return res.status(400).type("text").send("bad request body");
    }

    try {
      await new Promise((resolve, reject) => {
        const cmd = spawn("../sql/init.sh");
        cmd.stdout.pipe(process.stderr);
        cmd.stderr.pipe(process.stderr);
        cmd.on("exit", (code) => {
          resolve(code);
        });
        cmd.on("error", (err) => {
          reject(err);
        });
      });
    } catch (err) {
      console.error(`exec init.sh error: ${err}`);
      return res.status(500).send();
    }

    const db = await pool.getConnection();
    try {
      await db.query(
        "INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)",
        ["jia_service_url", request.jia_service_url]
      );
    } catch (err) {
      console.error(`db error: ${err}`);
      return res.status(500).send();
    } finally {
      db.release();
    }

    const initializeResponse: InitializeResponse = { language: "nodejs" };
    return res.status(200).json(initializeResponse);
  }
);

// POST /api/auth
// サインアップ・サインイン
app.post("/api/auth", async (req, res) => {
  const db = await pool.getConnection();
  try {
    const authHeader = req.headers.authorization ?? "";
    const token = authHeader.startsWith("Bearer ")
      ? authHeader.slice(7)
      : authHeader;

    let decoded: jwt.JwtPayload;
    try {
      decoded = jwt.verify(token, req.app.get("cert")) as jwt.JwtPayload;
    } catch (err) {
      return res.status(403).type("text").send("forbidden");
    }

    const jiaUserId = decoded["jia_user_id"];
    if (typeof jiaUserId !== "string") {
      return res.status(400).type("text").send("invalid JWT payload");
    }

    await db.query("INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)", [
      jiaUserId,
    ]);
    req.session = { jia_user_id: jiaUserId };

    return res.status(200).send();
  } catch (err) {
    console.error(`db error: ${err}`);
    return res.status(500).send();
  } finally {
    db.release();
  }
});

// POST /api/signout
// サインアウト
app.post("/api/signout", async (req, res) => {
  const db = await pool.getConnection();
  try {
    try {
      await getUserIdFromSession(req, db);
    } catch (err) {
      if (err instanceof ErrorWithStatus && err.status === 401) {
        return res.status(401).type("text").send("you are not signed in");
      }
      console.error(err);
      return res.status(500).send();
    }

    req.session = null;
    return res.status(200).send();
  } finally {
    db.release();
  }
});

// GET /api/user/me
// サインインしている自分自身の情報を取得
app.get("/api/user/me", async (req, res) => {
  const db = await pool.getConnection();
  try {
    let jiaUserId: string;
    try {
      jiaUserId = await getUserIdFromSession(req, db);
    } catch (err) {
      if (err instanceof ErrorWithStatus && err.status === 401) {
        return res.status(401).type("text").send("you are not signed in");
      }
      console.error(err);
      return res.status(500).send();
    }

    const getMeResponse: GetMeResponse = { jia_user_id: jiaUserId };
    return res.status(200).json(getMeResponse);
  } finally {
    db.release();
  }
});

// GET /api/isu
// ISUの一覧を取得
app.get("/api/isu", async (req, res) => {
  const db = await pool.getConnection();
  try {
    let jiaUserId: string;
    try {
      jiaUserId = await getUserIdFromSession(req, db);
    } catch (err) {
      if (err instanceof ErrorWithStatus && err.status === 401) {
        return res.status(401).type("text").send("you are not signed in");
      }
      console.error(err);
      return res.status(500).send();
    }

    await db.beginTransaction();

    const [isuList] = await db.query<Isu[]>(
      "SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC",
      [jiaUserId]
    );
    const responseList: Array<GetIsuListResponse> = [];
    for (const isu of isuList) {
      let foundLastCondition = true;
      const [[lastCondition]] = await db.query<IsuCondition[]>(
        "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1",
        [isu.jia_isu_uuid]
      );
      if (!lastCondition) {
        foundLastCondition = false;
      }
      let formattedCondition = undefined;
      if (foundLastCondition) {
        const [conditionLevel, err] = calculateConditionLevel(
          lastCondition.condition
        );
        if (err) {
          console.error(err);
          await db.rollback();
          return res.status(500).send();
        }
        formattedCondition = {
          jia_isu_uuid: lastCondition.jia_isu_uuid,
          isu_name: isu.name,
          timestamp: lastCondition.timestamp.getTime() / 1000,
          is_sitting: !!lastCondition.is_sitting,
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

    return res.status(200).json(responseList);
  } catch (err) {
    console.error(`db error: ${err}`);
    await db.rollback();
    return res.status(500).send();
  } finally {
    db.release();
  }
});

interface PostIsuRequest {
  jia_isu_uuid: string;
  isu_name: string;
}

// POST /api/isu
// ISUを登録
app.post(
  "/api/isu",
  (
    req: express.Request<Record<string, never>, unknown, PostIsuRequest>,
    res
  ) => {
    upload.single("image")(req, res, async (uploadErr) => {
      const db = await pool.getConnection();
      try {
        let jiaUserId: string;
        try {
          jiaUserId = await getUserIdFromSession(req, db);
        } catch (err) {
          if (err instanceof ErrorWithStatus && err.status === 401) {
            return res.status(401).type("text").send("you are not signed in");
          }
          console.error(err);
          return res.status(500).send();
        }

        const request = req.body;
        const jiaIsuUUID = request.jia_isu_uuid;
        const isuName = request.isu_name;
        if (uploadErr instanceof MulterError) {
          return res.send(400).send("bad format: icon");
        }

        const image = req.file
          ? req.file.buffer
          : await readFile(defaultIconFilePath);

        await db.beginTransaction();

        try {
          await db.query(
            "INSERT INTO `isu` (`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)",
            [jiaIsuUUID, isuName, image, jiaUserId]
          );
        } catch (err) {
          await db.rollback();
          if (err.errno === mysqlErrNumDuplicateEntry) {
            return res.status(409).type("text").send("duplicated: isu");
          } else {
            console.error(`db error: ${err}`);
            return res.status(500).send();
          }
        }

        const targetUrl = (await getJIAServiceUrl(db)) + "/api/activate";

        let isuFromJIA: { character: string };
        try {
          const response = await axios.post(
            targetUrl,
            {
              target_base_url: postIsuConditionTargetBaseURL,
              isu_uuid: jiaIsuUUID,
            },
            {
              validateStatus: (status) => status < 500,
            }
          );
          if (response.status !== 202) {
            console.error(
              `JIAService returned error: status code ${response.status}, message: ${response.data}`
            );
            await db.rollback();
            return res
              .status(response.status)
              .type("text")
              .send("JIAService returned error");
          }
          isuFromJIA = response.data;
        } catch (err) {
          console.error(`failed to request to JIAService: ${err}`);
          await db.rollback();
          return res.status(500).send();
        }

        await db.query(
          "UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?",
          [isuFromJIA.character, jiaIsuUUID]
        );
        const [[isu]] = await db.query<Isu[]>(
          "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
          [jiaUserId, jiaIsuUUID]
        );

        await db.commit();

        const isuResponse: IsuResponse = {
          id: isu.id,
          jia_isu_uuid: isu.jia_isu_uuid,
          name: isu.name,
          character: isu.character,
        };
        return res.status(201).send(isuResponse);
      } catch (err) {
        console.error(`db error: ${err}`);
        await db.rollback();
        return res.status(500).send();
      } finally {
        db.release();
      }
    });
  }
);

// GET /api/isu/:jia_isu_uuid
// ISUの情報を取得
app.get(
  "/api/isu/:jia_isu_uuid",
  async (req: express.Request<{ jia_isu_uuid: string }>, res) => {
    const db = await pool.getConnection();
    try {
      let jiaUserId: string;
      try {
        jiaUserId = await getUserIdFromSession(req, db);
      } catch (err) {
        if (err instanceof ErrorWithStatus && err.status === 401) {
          return res.status(401).type("text").send("you are not signed in");
        }
        console.error(err);
        return res.status(500).send();
      }

      const jiaIsuUUID = req.params.jia_isu_uuid;
      const [[isu]] = await db.query<Isu[]>(
        "SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        [jiaUserId, jiaIsuUUID]
      );
      if (!isu) {
        return res.status(404).type("text").send("not found: isu");
      }
      const isuResponse: IsuResponse = {
        id: isu.id,
        jia_isu_uuid: isu.jia_isu_uuid,
        name: isu.name,
        character: isu.character,
      };
      return res.status(200).json(isuResponse);
    } catch (err) {
      console.error(`db error: ${err}`);
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

// GET /api/isu/:jia_isu_uuid/icon
// ISUのアイコンを取得
app.get(
  "/api/isu/:jia_isu_uuid/icon",
  async (req: express.Request<{ jia_isu_uuid: string }>, res) => {
    const db = await pool.getConnection();
    try {
      let jiaUserId: string;
      try {
        jiaUserId = await getUserIdFromSession(req, db);
      } catch (err) {
        if (err instanceof ErrorWithStatus && err.status === 401) {
          return res.status(401).type("text").send("you are not signed in");
        }
        console.error(err);
        return res.status(500).send();
      }

      const jiaIsuUUID = req.params.jia_isu_uuid;
      const [[row]] = await db.query<(RowDataPacket & { image: Buffer })[]>(
        "SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        [jiaUserId, jiaIsuUUID]
      );
      if (!row) {
        return res.status(404).type("text").send("not found: isu");
      }
      return res.status(200).send(row.image);
    } catch (err) {
      console.error(`db error: ${err}`);
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

interface GetIsuGraphQuery extends qs.ParsedQs {
  datetime?: string;
}

// GET /api/isu/:jia_isu_uuid/graph
// ISUのコンディショングラフ描画のための情報を取得
app.get(
  "/api/isu/:jia_isu_uuid/graph",
  async (
    req: express.Request<
      { jia_isu_uuid: string },
      unknown,
      never,
      GetIsuGraphQuery
    >,
    res
  ) => {
    const db = await pool.getConnection();
    try {
      let jiaUserId: string;
      try {
        jiaUserId = await getUserIdFromSession(req, db);
      } catch (err) {
        if (err instanceof ErrorWithStatus && err.status === 401) {
          return res.status(401).type("text").send("you are not signed in");
        }
        console.error(err);
        return res.status(500).send();
      }

      const jiaIsuUUID = req.params.jia_isu_uuid;
      const datetimeStr = req.query.datetime;
      if (!datetimeStr) {
        return res.status(400).type("text").send("missing: datetime");
      }
      const datetime = parseInt(datetimeStr, 10);
      if (isNaN(datetime)) {
        return res.status(400).type("text").send("bad format: datetime");
      }
      const date = new Date(datetime * 1000);
      date.setMinutes(0, 0, 0);

      await db.beginTransaction();

      const [[{ cnt }]] = await db.query<(RowDataPacket & { cnt: number })[]>(
        "SELECT COUNT(*) AS `cnt` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?",
        [jiaUserId, jiaIsuUUID]
      );
      if (cnt === 0) {
        await db.rollback();
        return res.status(404).type("text").send("not found: isu");
      }
      const [getIsuGraphResponse, e] = await generateIsuGraphResponse(
        db,
        jiaIsuUUID,
        date
      );
      if (e) {
        console.error(e);
        await db.rollback();
        return res.status(500).send();
      }

      await db.commit();

      return res.status(200).json(getIsuGraphResponse);
    } catch (err) {
      console.error(`db error: ${err}`);
      await db.rollback();
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

async function generateIsuGraphResponse(
  db: mysql.Connection,
  jiaIsuUUID: string,
  graphDate: Date
): Promise<[GraphResponse[], Error?]> {
  const dataPoints: GraphDataPointWithInfo[] = [];
  let conditionsInThisHour = [];
  let timestampsInThisHour = [];
  let startTimeInThisHour = new Date(0);

  const [rows] = await db.query<IsuCondition[]>(
    "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC",
    [jiaIsuUUID]
  );
  for (const condition of rows) {
    const truncatedConditionTime = new Date(condition.timestamp);
    truncatedConditionTime.setMinutes(0, 0, 0);
    if (truncatedConditionTime.getTime() !== startTimeInThisHour.getTime()) {
      if (conditionsInThisHour.length > 0) {
        const [data, err] = calculateGraphDataPoint(conditionsInThisHour);
        if (err) {
          return [[], err];
        }
        dataPoints.push({
          jiaIsuUUID,
          startAt: startTimeInThisHour,
          data,
          conditionTimeStamps: timestampsInThisHour,
        });
      }
      startTimeInThisHour = truncatedConditionTime;
      conditionsInThisHour = [];
      timestampsInThisHour = [];
    }
    conditionsInThisHour.push(condition);
    timestampsInThisHour.push(condition.timestamp.getTime() / 1000);
  }

  if (conditionsInThisHour.length > 0) {
    const [data, err] = calculateGraphDataPoint(conditionsInThisHour);
    if (err) {
      return [[], err];
    }
    dataPoints.push({
      jiaIsuUUID,
      startAt: startTimeInThisHour,
      data,
      conditionTimeStamps: timestampsInThisHour,
    });
  }

  const endTime = new Date(graphDate.getTime() + 24 * 3600 * 1000);
  let startIndex = dataPoints.length;
  let endNextIndex = dataPoints.length;
  dataPoints.forEach((graph, i) => {
    if (startIndex === dataPoints.length && graph.startAt >= graphDate) {
      startIndex = i;
    }
    if (endNextIndex === dataPoints.length && graph.startAt > endTime) {
      endNextIndex = i;
    }
  });

  const filteredDataPoints: GraphDataPointWithInfo[] = [];
  if (startIndex < endNextIndex) {
    filteredDataPoints.push(...dataPoints.slice(startIndex, endNextIndex));
  }

  const responseList: GraphResponse[] = [];
  let index = 0;
  let thisTime = graphDate;

  while (thisTime < endTime) {
    let data = undefined;
    const timestamps: number[] = [];

    if (index < filteredDataPoints.length) {
      const dataWithInfo = filteredDataPoints[index];
      if (dataWithInfo.startAt.getTime() === thisTime.getTime()) {
        data = dataWithInfo.data;
        timestamps.push(...dataWithInfo.conditionTimeStamps);
        index++;
      }
    }

    responseList.push({
      start_at: thisTime.getTime() / 1000,
      end_at: thisTime.getTime() / 1000 + 3600,
      data,
      condition_timestamps: timestamps,
    });

    thisTime = new Date(thisTime.getTime() + 3600 * 1000);
  }

  return [responseList, undefined];
}

// 複数のISUのコンディションからグラフの一つのデータ点を計算
function calculateGraphDataPoint(
  isuConditions: IsuCondition[]
): [GraphDataPoint, Error?] {
  const conditionsCount: Record<string, number> = {
    is_broken: 0,
    is_dirty: 0,
    is_overweight: 0,
  };
  let rawScore = 0;
  isuConditions.forEach((condition) => {
    let badConditionsCount = 0;

    if (!isValidConditionFormat(condition.condition)) {
      return [{}, new Error("invalid condition format")];
    }

    condition.condition.split(",").forEach((condStr) => {
      const keyValue = condStr.split("=");

      const conditionName = keyValue[0];
      if (keyValue[1] === "true") {
        conditionsCount[conditionName] += 1;
        badConditionsCount++;
      }
    });

    if (badConditionsCount >= 3) {
      rawScore += scoreConditionLevelCritical;
    } else if (badConditionsCount >= 1) {
      rawScore += scoreConditionLevelWarning;
    } else {
      rawScore += scoreConditionLevelInfo;
    }
  });

  let sittingCount = 0;
  isuConditions.forEach((condition) => {
    if (condition.is_sitting) {
      sittingCount++;
    }
  });

  const isuConditionLength = isuConditions.length;
  const score = Math.trunc((rawScore * 100) / 3 / isuConditionLength);
  const sittingPercentage = Math.trunc(
    (sittingCount * 100) / isuConditionLength
  );
  const isBrokenPercentage = Math.trunc(
    (conditionsCount["is_broken"] * 100) / isuConditionLength
  );
  const isOverweightPercentage = Math.trunc(
    (conditionsCount["is_overweight"] * 100) / isuConditionLength
  );
  const isDirtyPercentage = Math.trunc(
    (conditionsCount["is_dirty"] * 100) / isuConditionLength
  );

  const dataPoint: GraphDataPoint = {
    score,
    percentage: {
      sitting: sittingPercentage,
      is_broken: isBrokenPercentage,
      is_overweight: isOverweightPercentage,
      is_dirty: isDirtyPercentage,
    },
  };
  return [dataPoint, undefined];
}

interface GetIsuConditionsQuery extends qs.ParsedQs {
  start_time: string;
  end_time: string;
  condition_level: string;
}

// GET /api/condition/:jia_isu_uuid
// ISUのコンディションを取得
app.get(
  "/api/condition/:jia_isu_uuid",
  async (
    req: express.Request<
      { jia_isu_uuid: string },
      unknown,
      never,
      GetIsuConditionsQuery
    >,
    res
  ) => {
    const db = await pool.getConnection();
    try {
      let jiaUserId: string;
      try {
        jiaUserId = await getUserIdFromSession(req, db);
      } catch (err) {
        if (err instanceof ErrorWithStatus && err.status === 401) {
          return res.status(401).type("text").send("you are not signed in");
        }
        console.error(err);
        return res.status(500).send();
      }

      const jiaIsuUUID = req.params.jia_isu_uuid;

      const endTimeInt = parseInt(req.query.end_time, 10);
      if (isNaN(endTimeInt)) {
        return res.status(400).type("text").send("bad format: end_time");
      }
      const endTime = new Date(endTimeInt * 1000);
      if (!req.query.condition_level) {
        return res.status(400).type("text").send("missing: condition_level");
      }
      const conditionLevel = new Set(req.query.condition_level.split(","));

      const startTimeStr = req.query.start_time;
      let startTime = new Date(0);
      if (startTimeStr) {
        const startTimeInt = parseInt(startTimeStr, 10);
        if (isNaN(startTimeInt)) {
          return res.status(400).type("text").send("bad format: start_time");
        }
        startTime = new Date(startTimeInt * 1000);
      }

      const [[row]] = await db.query<(RowDataPacket & { name: string })[]>(
        "SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?",
        [jiaIsuUUID, jiaUserId]
      );
      if (!row) {
        return res.status(404).type("text").send("not found: isu");
      }

      const conditionResponse: GetIsuConditionResponse[] =
        await getIsuConditions(
          db,
          jiaIsuUUID,
          endTime,
          conditionLevel,
          startTime,
          conditionLimit,
          row.name
        );
      res.status(200).json(conditionResponse);
    } catch (err) {
      console.error(`db error: ${err}`);
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

// ISUのコンディションをDBから取得
async function getIsuConditions(
  db: mysql.Connection,
  jiaIsuUUID: string,
  endTime: Date,
  conditionLevel: Set<string>,
  startTime: Date,
  limit: number,
  isuName: string
): Promise<GetIsuConditionResponse[]> {
  const [conditions] =
    startTime.getTime() === 0
      ? await db.query<IsuCondition[]>(
          "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?" +
            "	AND `timestamp` < ?" +
            "	ORDER BY `timestamp` DESC",
          [jiaIsuUUID, endTime]
        )
      : await db.query<IsuCondition[]>(
          "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?" +
            "	AND `timestamp` < ?" +
            "	AND ? <= `timestamp`" +
            "	ORDER BY `timestamp` DESC",
          [jiaIsuUUID, endTime, startTime]
        );

  let conditionsResponse: GetIsuConditionResponse[] = [];
  conditions.forEach((condition) => {
    const [cLevel, err] = calculateConditionLevel(condition.condition);
    if (err) {
      return;
    }
    if (conditionLevel.has(cLevel)) {
      conditionsResponse.push({
        jia_isu_uuid: condition.jia_isu_uuid,
        isu_name: isuName,
        timestamp: condition.timestamp.getTime() / 1000,
        is_sitting: !!condition.is_sitting,
        condition: condition.condition,
        condition_level: cLevel,
        message: condition.message,
      });
    }
  });

  if (conditionsResponse.length > limit) {
    conditionsResponse = conditionsResponse.slice(0, limit);
  }

  return conditionsResponse;
}

// ISUのコンディションの文字列からコンディションレベルを計算
function calculateConditionLevel(condition: string): [string, Error?] {
  let conditionLevel: string;
  const warnCount = (() => {
    let count = 0;
    let pos = 0;
    while (pos !== -1) {
      pos = condition.indexOf("=true", pos);
      if (pos >= 0) {
        count += 1;
        pos += 5;
      }
    }
    return count;
  })();
  switch (warnCount) {
    case 0:
      conditionLevel = conditionLevelInfo;
      break;
    case 1: // fallthrough
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

// GET /api/trend
// ISUの性格毎の最新のコンディション情報
app.get("/api/trend", async (req, res) => {
  const db = await pool.getConnection();
  try {
    const [characterList] = await db.query<
      (RowDataPacket & { character: string })[]
    >("SELECT `character` FROM `isu` GROUP BY `character`");

    const trendResponse: TrendResponse[] = [];

    for (const character of characterList) {
      const [isuList] = await db.query<Isu[]>(
        "SELECT * FROM `isu` WHERE `character` = ?",
        [character.character]
      );

      const characterInfoIsuConditions = [];
      const characterWarningIsuConditions = [];
      const characterCriticalIsuConditions = [];
      for (const isu of isuList) {
        const [conditions] = await db.query<IsuCondition[]>(
          "SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY timestamp DESC",
          [isu.jia_isu_uuid]
        );

        if (conditions.length > 0) {
          const isuLastCondition = conditions[0];
          const [conditionLevel, err] = calculateConditionLevel(
            isuLastCondition.condition
          );
          if (err) {
            console.error(err);
            return res.status(500).send();
          }
          const trendCondition: TrendCondition = {
            isu_id: isu.id,
            timestamp: isuLastCondition.timestamp.getTime() / 1000,
          };
          switch (conditionLevel) {
            case "info":
              characterInfoIsuConditions.push(trendCondition);
              break;
            case "warning":
              characterWarningIsuConditions.push(trendCondition);
              break;
            case "critical":
              characterCriticalIsuConditions.push(trendCondition);
              break;
          }
        }
      }

      characterInfoIsuConditions.sort((a, b) => b.timestamp - a.timestamp);
      characterWarningIsuConditions.sort((a, b) => b.timestamp - a.timestamp);
      characterCriticalIsuConditions.sort((a, b) => b.timestamp - a.timestamp);
      trendResponse.push({
        character: character.character,
        info: characterInfoIsuConditions,
        warning: characterWarningIsuConditions,
        critical: characterCriticalIsuConditions,
      });
    }

    return res.status(200).json(trendResponse);
  } catch (err) {
    console.error(`db error: ${err}`);
    return res.status(500).send();
  } finally {
    db.release();
  }
});

interface PostIsuConditionRequest {
  is_sitting: boolean;
  condition: string;
  message: string;
  timestamp: number;
}

function isValidPostIsuConditionRequest(
  body: PostIsuConditionRequest[]
): body is PostIsuConditionRequest[] {
  return (
    Array.isArray(body) &&
    body.every((data) => {
      return (
        typeof data.is_sitting === "boolean" &&
        typeof data.condition === "string" &&
        typeof data.message === "string" &&
        typeof data.timestamp === "number"
      );
    })
  );
}

// POST /api/condition/:jia_isu_uuid
// ISUからのコンディションを受け取る
app.post(
  "/api/condition/:jia_isu_uuid",
  async (
    req: express.Request<
      { jia_isu_uuid: string },
      unknown,
      PostIsuConditionRequest[]
    >,
    res
  ) => {
    // TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
    const dropProbability = 0.9;
    if (Math.random() <= dropProbability) {
      console.warn("drop post isu condition request");
      return res.status(202).send();
    }

    const db = await pool.getConnection();
    try {
      const jiaIsuUUID = req.params.jia_isu_uuid;

      const request = req.body;
      if (!isValidPostIsuConditionRequest(request) || request.length === 0) {
        return res.status(400).type("text").send("bad request body");
      }

      await db.beginTransaction();

      const [[{ cnt }]] = await db.query<(RowDataPacket & { cnt: number })[]>(
        "SELECT COUNT(*) AS `cnt` FROM `isu` WHERE `jia_isu_uuid` = ?",
        [jiaIsuUUID]
      );
      if (cnt === 0) {
        await db.rollback();
        return res.status(404).type("text").send("not found: isu");
      }

      for (const cond of request) {
        const timestamp = new Date(cond.timestamp * 1000);

        if (!isValidConditionFormat(cond.condition)) {
          await db.rollback();
          return res.status(400).type("text").send("bad request body");
        }

        await db.query(
          "INSERT INTO `isu_condition`" +
            "	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)" +
            "	VALUES (?, ?, ?, ?, ?)",
          [jiaIsuUUID, timestamp, cond.is_sitting, cond.condition, cond.message]
        );
      }

      await db.commit();

      return res.status(202).send();
    } catch (err) {
      console.error(`db error: ${err}`);
      await db.rollback();
      return res.status(500).send();
    } finally {
      db.release();
    }
  }
);

// ISUのコンディションの文字列がcsv形式になっているか検証
function isValidConditionFormat(condition: string): boolean {
  const keys = ["is_dirty=", "is_overweight=", "is_broken="];
  const valueTrue = "true";
  const valueFalse = "false";

  let idxCondStr = 0;

  for (const [idxKeys, key] of keys.entries()) {
    if (!condition.slice(idxCondStr).startsWith(key)) {
      return false;
    }
    idxCondStr += key.length;

    if (condition.slice(idxCondStr).startsWith(valueTrue)) {
      idxCondStr += valueTrue.length;
    } else if (condition.slice(idxCondStr).startsWith(valueFalse)) {
      idxCondStr += valueFalse.length;
    } else {
      return false;
    }

    if (idxKeys < keys.length - 1) {
      if (condition[idxCondStr] !== ",") {
        return false;
      }
      idxCondStr++;
    }
  }

  return idxCondStr === condition.length;
}

[
  "/",
  "/isu/:jia_isu_uuid",
  "/isu/:jia_isu_uuid/condition",
  "/isu/:jia_isu_uuid/graph",
  "/register",
].forEach((frontendPath) => {
  app.get(frontendPath, (_req, res) => {
    res.sendFile(path.resolve("../public", "index.html"));
  });
});

app.listen(parseInt(process.env["SERVER_APP_PORT"] ?? "3000", 10));
