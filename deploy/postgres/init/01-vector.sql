-- 应用启动时也会执行 CREATE EXTENSION，这里提前创建以便 initdb 阶段就绪。
CREATE EXTENSION IF NOT EXISTS vector;
