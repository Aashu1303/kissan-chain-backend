require("dotenv").config();
const { ethers } = require("ethers");
const fs = require("fs");

// Load environment variables
const { HARDHAT_NETWORK_URL, PRIVATE_KEY, CONTRACT_ADDRESS } = process.env;

// Load contract ABI
const contractABI = JSON.parse(
    fs.readFileSync("./Auction.json", "utf8")
);

// Connect to Hardhat local network

const provider = new ethers.JsonRpcProvider(HARDHAT_NETWORK_URL);
const wallet = new ethers.Wallet(PRIVATE_KEY, provider);
const auctionContract = new ethers.Contract(CONTRACT_ADDRESS, contractABI.abi, wallet);

// Create a listing
async function createListing() {
    const tx = await auctionContract.createListing(
        1,
        "Painting",
        "QmImageHashExample",
        ethers.parseEther("0.1"), // Minimum bid
        3600, // Duration in seconds
        wallet.address
    );
    console.log("Transaction hash:", tx.hash);
    await tx.wait();
    console.log("Listing created.");
}

// Fetch all listings
async function fetchAllListings() {
    const listings = await auctionContract.fetchAllListings();
    console.log("All listings:", listings);
}

// Main function to execute tasks
async function main() {
    await createListing();
    await fetchAllListings();
}

main().catch(console.error);
