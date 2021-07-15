import express from "express";

export const app = express();

app.listen(process.env.PORT ?? 3000, () => {});
