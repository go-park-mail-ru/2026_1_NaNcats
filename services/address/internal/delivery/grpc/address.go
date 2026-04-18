package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Маппер из Protobuf в доменную структуру
func mapPBToDomainAddress(a *pb.Address) domain.Address {
	return domain.Address{
		PublicID: a.PublicId,
		Location: domain.Location{
			AddressText: a.Location.AddressText,
			Latitude:    a.Location.Latitude,
			Longitude:   a.Location.Longitude,
		},
		Apartment:      a.Apartment,
		Entrance:       a.Entrance,
		Floor:          a.Floor,
		DoorCode:       a.DoorCode,
		CourierComment: a.CourierComment,
		Label:          a.Label,
	}
}

// Маппер из доменной структуры в Protobuf
func mapDomainToPBAddress(a domain.Address) *pb.Address {
	return &pb.Address{
		PublicId: a.PublicID,
		Location: &pb.Location{
			AddressText: a.Location.AddressText,
			Latitude:    a.Location.Latitude,
			Longitude:   a.Location.Longitude,
		},
		Apartment:      a.Apartment,
		Entrance:       a.Entrance,
		Floor:          a.Floor,
		DoorCode:       a.DoorCode,
		CourierComment: a.CourierComment,
		Label:          a.Label,
	}
}

type AddressHandler struct {
	pb.UnimplementedAddressServiceServer
	usecase usecase.AddressUseCase
}

func NewAddressHandler(uc usecase.AddressUseCase) *AddressHandler {
	return &AddressHandler{
		usecase: uc,
	}
}

func (h *AddressHandler) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.AddAddressResponse, error) {
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	domainAddr := mapPBToDomainAddress(req.Address)

	addressPublicID, err := h.usecase.AddAddress(ctx, req.UserId, domainAddr, req.IdempotencyKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add address: %v", err)
	}

	return &pb.AddAddressResponse{
		AddressPublicId: addressPublicID,
	}, nil
}

func (h *AddressHandler) GetMyAddresses(ctx context.Context, req *pb.GetMyAddressesRequest) (*pb.GetMyAddressesResponse, error) {
	domainAddresses, err := h.usecase.GetMyAddresses(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get addresses: %v", err)
	}

	pbAddreses := make([]*pb.Address, 0, len(domainAddresses))
	for _, addr := range domainAddresses {
		pbAddreses = append(pbAddreses, mapDomainToPBAddress(addr))
	}

	return &pb.GetMyAddressesResponse{
		Addresses: pbAddreses,
	}, nil
}

func (h *AddressHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	err := h.usecase.DeleteAddress(ctx, req.UserId, req.AddressPublicId, req.IdempotencyKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete address: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (h *AddressHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*emptypb.Empty, error) {
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	domainAddress := mapPBToDomainAddress(req.Address)

	err := h.usecase.UpdateAddress(ctx, req.UserId, domainAddress, req.IdempotencyKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update address: %v", err)
	}

	return &emptypb.Empty{}, nil
}
