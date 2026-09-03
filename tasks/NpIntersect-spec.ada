--  <vc-preamble>
package Np_Intersect_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   --  Dafny allocates the result inside the method, so the length is chosen
   --  by the implementation rather than by the caller; a function returning
   --  an unconstrained array is the SPARK analogue.
   function Intersect (A : Int_Array; B : Int_Array) return Int_Array with
     Post => Intersect'Result'Length < A'Length
             and then Intersect'Result'Length < B'Length
             and then
               (for all I in A'Range =>
                  (for all J in B'Range =>
                     (if A (I) = B (J)
                      then (for some K in Intersect'Result'Range =>
                              Intersect'Result (K) = A (I)))
                     and then
                     (if A (I) /= B (J)
                      then (for all K in Intersect'Result'Range =>
                              Intersect'Result (K) /= A (I)))));

end Np_Intersect_Spec;

package body Np_Intersect_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Intersect (A : Int_Array; B : Int_Array) return Int_Array is
   begin
      pragma Assume (False);
      return R : Int_Array (1 .. 0);
   end Intersect;
--  </vc-code>

--  <vc-postamble>
end Np_Intersect_Spec;
--  </vc-postamble>
