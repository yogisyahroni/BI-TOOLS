ALTER TABLE "Dashboard" DROP CONSTRAINT "Dashboard_collectionId_fkey";
ALTER TABLE "Dashboard"
ADD CONSTRAINT "Dashboard_collectionId_fkey" FOREIGN KEY ("collectionId") REFERENCES collections(id);
ALTER TABLE "Dashboard" DROP CONSTRAINT "Dashboard_userId_fkey";
ALTER TABLE "Dashboard"
ADD CONSTRAINT "Dashboard_userId_fkey" FOREIGN KEY ("userId") REFERENCES users(id);